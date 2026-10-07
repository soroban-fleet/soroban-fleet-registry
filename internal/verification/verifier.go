package verification

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/soroban-fleet/soroban-fleet-registry/internal/cap85"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/fleet"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/ingest"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/rpc"
)

// Verifier defines the interface for running and querying fleet integrity verifications.
type Verifier interface {
	VerifyFleet(ctx context.Context, fleetID FleetID, expectedHash []byte) (VerificationResult, error)
	ListVerifications(ctx context.Context, fleetID FleetID, limit, offset int) ([]VerificationResult, int64, error)
	GetLatestVerification(ctx context.Context, fleetID FleetID) (*VerificationResult, error)
}

// Engine implements Verifier using live resolution and persistent storage.
type Engine struct {
	db          *sql.DB
	fleetRepo   fleet.Repository
	checkpoints ingest.CheckpointStore
	resolver    cap85.Resolver
	rpcClient   rpc.Client
}

// NewEngine creates a new verification Engine.
func NewEngine(
	db *sql.DB,
	fleetRepo fleet.Repository,
	checkpoints ingest.CheckpointStore,
	resolver cap85.Resolver,
	rpcClient rpc.Client,
) *Engine {
	return &Engine{
		db:          db,
		fleetRepo:   fleetRepo,
		checkpoints: checkpoints,
		resolver:    resolver,
		rpcClient:   rpcClient,
	}
}

// VerifyFleet verifies that all active members in a fleet resolve to the expected WASM hash.
func (e *Engine) VerifyFleet(ctx context.Context, fleetID FleetID, expectedHash []byte) (VerificationResult, error) {
	now := time.Now().UTC()

	// 1. Input validation
	if len(expectedHash) != 32 {
		return VerificationResult{}, fmt.Errorf("expected WASM hash must be exactly 32 bytes, got %d", len(expectedHash))
	}
	if _, err := cap85.AddressToScAddress(fleetID.Owner); err != nil {
		return VerificationResult{}, fmt.Errorf("%w: %s", cap85.ErrInvalidContractID, fleetID.Owner)
	}

	// 2. Load active members
	members, err := e.fleetRepo.GetActiveMembers(ctx, fleetID)
	if err != nil {
		return VerificationResult{}, fmt.Errorf("load active members for fleet %v: %w", fleetID, err)
	}

	// 3. Obtain current indexer checkpoint and latest network ledger
	var indexedThrough uint32
	if e.checkpoints != nil {
		cp, err := e.checkpoints.GetCheckpoint(ctx, ingest.DefaultStreamName)
		if err == nil {
			indexedThrough = cp
		}
	}

	var latestLedger uint32 = indexedThrough
	if e.rpcClient != nil {
		latest, err := e.rpcClient.GetLatestLedger(ctx)
		if err == nil && latest.Sequence > 0 {
			latestLedger = latest.Sequence
		}
	}

	// 4. Resolve current executable state for each member
	var total int64 = int64(len(members))
	var matching int64
	var mismatching int64
	var missing int64
	var brokenRef bool

	if len(members) > 0 && e.resolver != nil {
		contractIDs := make([]string, len(members))
		for i, m := range members {
			contractIDs[i] = m.ContractID
		}

		resolvedMap, err := e.resolver.ResolveContractExecutablesBatch(ctx, contractIDs)
		if err != nil {
			if errors.Is(err, cap85.ErrBrokenReference) {
				brokenRef = true
			}
		}

		for _, m := range members {
			res, ok := resolvedMap[m.ContractID]
			if !ok {
				// Try single resolution fallback
				singleRes, sErr := e.resolver.ResolveContractExecutable(ctx, m.ContractID)
				if sErr != nil {
					missing++
					if errors.Is(sErr, cap85.ErrBrokenReference) {
						brokenRef = true
					}
					continue
				}
				res = singleRes
			}

			if len(res.WASMHash) == 0 {
				missing++
				continue
			}

			if bytes.Equal(res.WASMHash, expectedHash) {
				matching++
			} else {
				mismatching++
			}
		}
	}

	// 5. Determine status according to Section 17 rules
	var status VerificationStatus

	switch {
	case brokenRef:
		status = StatusBrokenReference

	case mismatching > 0:
		status = StatusDrift

	case missing > 0:
		status = StatusIncomplete

	case indexedThrough < latestLedger:
		// Indexer has not reached the network height required for a complete assertion
		status = StatusIncomplete

	case total == 0:
		// Critical safety rule: Never return HEALTHY simply because no mismatch was found
		status = StatusIncomplete

	case matching == total && indexedThrough >= latestLedger:
		status = StatusHealthy

	default:
		status = StatusUnknown
	}

	result := VerificationResult{
		FleetID:            fleetID,
		ExpectedWASMHash:   expectedHash,
		TotalMembers:       total,
		MatchingMembers:    matching,
		MismatchingMembers: mismatching,
		MissingMembers:     missing,
		IndexedThrough:     indexedThrough,
		Status:             status,
		VerifiedAt:         now,
	}

	// 6. Persist verification result if database is configured
	if e.db != nil {
		query := `
			INSERT INTO fleet_verifications (
				owner_address, tag, expected_wasm_hash,
				total_members, matching_members, mismatching_members, missing_members,
				status, indexed_through, verified_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
			RETURNING id
		`
		var id int64
		err := e.db.QueryRowContext(
			ctx, query,
			fleetID.Owner, fleetID.Tag, expectedHash,
			total, matching, mismatching, missing,
			string(status), indexedThrough, now,
		).Scan(&id)
		if err != nil {
			return result, fmt.Errorf("persist verification result: %w", err)
		}
		result.ID = id
	}

	return result, nil
}

// ListVerifications returns paginated past verifications for a fleet.
func (e *Engine) ListVerifications(ctx context.Context, fleetID FleetID, limit, offset int) ([]VerificationResult, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	var total int64
	countQuery := "SELECT COUNT(*) FROM fleet_verifications WHERE owner_address = $1 AND tag = $2"
	if err := e.db.QueryRowContext(ctx, countQuery, fleetID.Owner, fleetID.Tag).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count verifications: %w", err)
	}

	query := `
		SELECT id, owner_address, tag, expected_wasm_hash,
		       total_members, matching_members, mismatching_members, missing_members,
		       status, indexed_through, verified_at
		FROM fleet_verifications
		WHERE owner_address = $1 AND tag = $2
		ORDER BY verified_at DESC, id DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := e.db.QueryContext(ctx, query, fleetID.Owner, fleetID.Tag, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list verifications: %w", err)
	}
	defer rows.Close()

	var results []VerificationResult
	for rows.Next() {
		var vr VerificationResult
		var statusStr string
		var indexedInt int64
		if err := rows.Scan(
			&vr.ID,
			&vr.FleetID.Owner,
			&vr.FleetID.Tag,
			&vr.ExpectedWASMHash,
			&vr.TotalMembers,
			&vr.MatchingMembers,
			&vr.MismatchingMembers,
			&vr.MissingMembers,
			&statusStr,
			&indexedInt,
			&vr.VerifiedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan verification: %w", err)
		}
		vr.Status = VerificationStatus(statusStr)
		vr.IndexedThrough = uint32(indexedInt)
		results = append(results, vr)
	}

	return results, total, rows.Err()
}

// GetLatestVerification returns the most recent verification record for a fleet.
func (e *Engine) GetLatestVerification(ctx context.Context, fleetID FleetID) (*VerificationResult, error) {
	query := `
		SELECT id, owner_address, tag, expected_wasm_hash,
		       total_members, matching_members, mismatching_members, missing_members,
		       status, indexed_through, verified_at
		FROM fleet_verifications
		WHERE owner_address = $1 AND tag = $2
		ORDER BY verified_at DESC, id DESC
		LIMIT 1
	`
	row := e.db.QueryRowContext(ctx, query, fleetID.Owner, fleetID.Tag)

	var vr VerificationResult
	var statusStr string
	var indexedInt int64
	err := row.Scan(
		&vr.ID,
		&vr.FleetID.Owner,
		&vr.FleetID.Tag,
		&vr.ExpectedWASMHash,
		&vr.TotalMembers,
		&vr.MatchingMembers,
		&vr.MismatchingMembers,
		&vr.MissingMembers,
		&statusStr,
		&indexedInt,
		&vr.VerifiedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get latest verification: %w", err)
	}
	vr.Status = VerificationStatus(statusStr)
	vr.IndexedThrough = uint32(indexedInt)
	return &vr, nil
}
