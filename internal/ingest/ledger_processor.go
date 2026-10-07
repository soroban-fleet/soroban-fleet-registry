package ingest

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/soroban-fleet/soroban-fleet-registry/internal/cap85"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/fleet"
	"github.com/stellar/go-stellar-sdk/xdr"
)

// LedgerProcessingStats holds metrics about a processed ledger cycle.
type LedgerProcessingStats struct {
	Ledger           uint32        `json:"ledger"`
	Duration         time.Duration `json:"duration_ms"`
	ChangesProcessed int           `json:"changes_processed"`
	FleetsUpdated    int           `json:"fleets_updated"`
	MembersUpdated   int           `json:"members_updated"`
	ReleasesDetected int           `json:"releases_detected"`
	Errors           []string      `json:"errors,omitempty"`
}

// ReleaseRecorder is an interface implemented by history/releases to record releases during ingestion.
type ReleaseRecorder interface {
	RecordReleaseTx(ctx context.Context, tx *sql.Tx, owner, tag string, oldWasm, newWasm []byte, ledger uint32, txHash string, observedAt time.Time) error
}

// LedgerProcessor applies ledger changes to the database within a transaction.
type LedgerProcessor struct {
	fleetRepo  fleet.Repository
	membership *fleet.MembershipManager
	resolver   cap85.Resolver
	releases   ReleaseRecorder
}

// NewLedgerProcessor creates a new LedgerProcessor.
func NewLedgerProcessor(
	fleetRepo fleet.Repository,
	membership *fleet.MembershipManager,
	resolver cap85.Resolver,
	releases ReleaseRecorder,
) *LedgerProcessor {
	return &LedgerProcessor{
		fleetRepo:  fleetRepo,
		membership: membership,
		resolver:   resolver,
		releases:   releases,
	}
}

// ProcessLedgerTx processes all changes for a single ledger inside an active database transaction.
func (p *LedgerProcessor) ProcessLedgerTx(
	ctx context.Context,
	tx *sql.Tx,
	changes LedgerChanges,
) (LedgerProcessingStats, error) {
	start := time.Now()
	stats := LedgerProcessingStats{
		Ledger: changes.Sequence,
	}

	observedAt := time.Unix(changes.CloseTimeUnix, 0).UTC()
	if changes.CloseTimeUnix == 0 {
		observedAt = time.Now().UTC()
	}

	knownWasm := make(map[fleet.FleetID][]byte)

	// 1. Process tag executable changes (owner contract updated WASM hash for tag)
	for _, tc := range changes.TagChanges {
		stats.ChangesProcessed++
		fID := fleet.FleetID{Owner: tc.OwnerAddress, Tag: tc.Tag}
		knownWasm[fID] = tc.NewWASMHash

		// Upsert fleet with new WASM hash
		existing, err := p.fleetRepo.GetFleetTx(ctx, tx, fID)
		firstSeen := changes.Sequence
		var oldWasm []byte

		if err == nil && existing != nil {
			firstSeen = existing.FirstSeenLedger
			oldWasm = existing.CurrentWASMHash
		} else {
			oldWasm = tc.OldWASMHash
		}

		flt := &fleet.Fleet{
			ID:                fID,
			CurrentWASMHash:   tc.NewWASMHash,
			MemberCount:       0,
			FirstSeenLedger:   firstSeen,
			LastSeenLedger:    changes.Sequence,
			LastIndexedLedger: changes.Sequence,
			CreatedAt:         observedAt,
			UpdatedAt:         observedAt,
		}

		if err := p.fleetRepo.UpsertFleetTx(ctx, tx, flt); err != nil {
			return stats, fmt.Errorf("upsert fleet from tag change %v: %w", fID, err)
		}
		stats.FleetsUpdated++

		// Record release if configured
		if p.releases != nil {
			if err := p.releases.RecordReleaseTx(ctx, tx, tc.OwnerAddress, tc.Tag, oldWasm, tc.NewWASMHash, changes.Sequence, tc.TxHash, observedAt); err != nil {
				return stats, fmt.Errorf("record release for %v: %w", fID, err)
			}
			stats.ReleasesDetected++
		}
	}

	// 2. Process contract instance changes
	for _, cc := range changes.ContractChanges {
		stats.ChangesProcessed++

		// Case A: New executable is an external ref
		if cc.NewExec != nil && cc.NewExec.Type == xdr.ContractExecutableTypeContractExecutableExternalRef {
			extRef, isExt, err := cap85.DecodeExternalExecutableRef(*cc.NewExec)
			if err != nil {
				stats.Errors = append(stats.Errors, fmt.Sprintf("contract %s: %v", cc.ContractID, err))
				continue
			}
			if !isExt {
				continue
			}

			fID := extRef.FleetID()

			// Check if fleet exists or create it
			f, err := p.fleetRepo.GetFleetTx(ctx, tx, fID)
			var currentWasm []byte
			firstSeen := changes.Sequence

			if err == nil && f != nil {
				firstSeen = f.FirstSeenLedger
				currentWasm = f.CurrentWASMHash
			}
			if len(currentWasm) == 0 {
				if kw, ok := knownWasm[fID]; ok {
					currentWasm = kw
				}
			}

			// If current WASM is not known in database yet, resolve it via resolver if available
			if len(currentWasm) == 0 && p.resolver != nil {
				if res, err := p.resolver.ResolveContractExecutable(ctx, cc.ContractID); err == nil && len(res.WASMHash) == 32 {
					currentWasm = res.WASMHash
				}
			}

			flt := &fleet.Fleet{
				ID:                fID,
				CurrentWASMHash:   currentWasm,
				FirstSeenLedger:   firstSeen,
				LastSeenLedger:    changes.Sequence,
				LastIndexedLedger: changes.Sequence,
				CreatedAt:         observedAt,
				UpdatedAt:         observedAt,
			}
			if err := p.fleetRepo.UpsertFleetTx(ctx, tx, flt); err != nil {
				return stats, fmt.Errorf("upsert fleet for member %s: %w", cc.ContractID, err)
			}
			stats.FleetsUpdated++

			// Register or update member in fleet
			if err := p.membership.RegisterOrUpdateMemberTx(ctx, tx, cc.ContractID, fID, currentWasm, changes.Sequence); err != nil {
				return stats, fmt.Errorf("register member %s: %w", cc.ContractID, err)
			}
			stats.MembersUpdated++

		} else if cc.OldExec != nil && cc.OldExec.Type == xdr.ContractExecutableTypeContractExecutableExternalRef {
			// Case B: Old executable was an external ref, but new executable is not (or removed)
			if err := p.membership.DeactivateMemberTx(ctx, tx, cc.ContractID, changes.Sequence); err != nil {
				return stats, fmt.Errorf("deactivate member %s: %w", cc.ContractID, err)
			}
			stats.MembersUpdated++
		}
	}

	stats.Duration = time.Since(start)
	return stats, nil
}
