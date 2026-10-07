package history

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/soroban-fleet/soroban-fleet-registry/internal/cap85"
)

var (
	// ErrReleaseNotFound indicates no release exists matching criteria.
	ErrReleaseNotFound = errors.New("release not found")
)

// FleetID uniquely identifies a fleet.
type FleetID = cap85.FleetID

// FleetRelease represents an observed upgrade or executable release for a fleet.
type FleetRelease struct {
	ID          int64     `json:"id"`
	FleetID     FleetID   `json:"fleet_id"`
	OldWASMHash []byte    `json:"old_wasm_hash,omitempty"`
	NewWASMHash []byte    `json:"new_wasm_hash"`
	Ledger      uint32    `json:"ledger"`
	TxHash      string    `json:"tx_hash"`
	ObservedAt  time.Time `json:"observed_at"`
}

// Repository defines data access for fleet release history.
type Repository interface {
	RecordReleaseTx(ctx context.Context, tx *sql.Tx, owner, tag string, oldWasm, newWasm []byte, ledger uint32, txHash string, observedAt time.Time) error
	ListReleases(ctx context.Context, fleetID FleetID, limit, offset int) ([]FleetRelease, int64, error)
	GetLatestRelease(ctx context.Context, fleetID FleetID) (*FleetRelease, error)
}

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new history PostgresRepository.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// RecordReleaseTx inserts a new release record within an active transaction if the WASM hash changed and not already recorded.
func (r *PostgresRepository) RecordReleaseTx(
	ctx context.Context,
	tx *sql.Tx,
	owner, tag string,
	oldWasm, newWasm []byte,
	ledger uint32,
	txHash string,
	observedAt time.Time,
) error {
	// Only record if WASM hash actually changed
	if len(oldWasm) > 0 && bytes.Equal(oldWasm, newWasm) {
		return nil
	}

	// Idempotency: verify this release was not already recorded in this ledger/tx
	var exists bool
	checkQuery := `
		SELECT EXISTS (
			SELECT 1 FROM fleet_releases
			WHERE owner_address = $1 AND tag = $2 AND ledger = $3 AND tx_hash = $4
		)
	`
	if err := tx.QueryRowContext(ctx, checkQuery, owner, tag, ledger, txHash).Scan(&exists); err != nil {
		return fmt.Errorf("check existing release: %w", err)
	}
	if exists {
		return nil
	}

	insertQuery := `
		INSERT INTO fleet_releases (
			owner_address, tag, old_wasm_hash, new_wasm_hash,
			ledger, tx_hash, observed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := tx.ExecContext(ctx, insertQuery, owner, tag, oldWasm, newWasm, ledger, txHash, observedAt)
	if err != nil {
		return fmt.Errorf("insert release for %s:%s at ledger %d: %w", owner, tag, ledger, err)
	}

	return nil
}

// ListReleases returns paginated historical releases for a fleet.
func (r *PostgresRepository) ListReleases(ctx context.Context, fleetID FleetID, limit, offset int) ([]FleetRelease, int64, error) {
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
	countQuery := "SELECT COUNT(*) FROM fleet_releases WHERE owner_address = $1 AND tag = $2"
	if err := r.db.QueryRowContext(ctx, countQuery, fleetID.Owner, fleetID.Tag).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count releases: %w", err)
	}

	query := `
		SELECT id, owner_address, tag, old_wasm_hash, new_wasm_hash,
		       ledger, tx_hash, observed_at
		FROM fleet_releases
		WHERE owner_address = $1 AND tag = $2
		ORDER BY ledger DESC, id DESC
		LIMIT $3 OFFSET $4
	`
	rows, err := r.db.QueryContext(ctx, query, fleetID.Owner, fleetID.Tag, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query releases: %w", err)
	}
	defer rows.Close()

	var releases []FleetRelease
	for rows.Next() {
		var rel FleetRelease
		var ledgerInt int64
		if err := rows.Scan(
			&rel.ID,
			&rel.FleetID.Owner,
			&rel.FleetID.Tag,
			&rel.OldWASMHash,
			&rel.NewWASMHash,
			&ledgerInt,
			&rel.TxHash,
			&rel.ObservedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan release: %w", err)
		}
		rel.Ledger = uint32(ledgerInt)
		releases = append(releases, rel)
	}

	return releases, total, rows.Err()
}

// GetLatestRelease returns the most recent release for a fleet.
func (r *PostgresRepository) GetLatestRelease(ctx context.Context, fleetID FleetID) (*FleetRelease, error) {
	query := `
		SELECT id, owner_address, tag, old_wasm_hash, new_wasm_hash,
		       ledger, tx_hash, observed_at
		FROM fleet_releases
		WHERE owner_address = $1 AND tag = $2
		ORDER BY ledger DESC, id DESC
		LIMIT 1
	`
	row := r.db.QueryRowContext(ctx, query, fleetID.Owner, fleetID.Tag)

	var rel FleetRelease
	var ledgerInt int64
	err := row.Scan(
		&rel.ID,
		&rel.FleetID.Owner,
		&rel.FleetID.Tag,
		&rel.OldWASMHash,
		&rel.NewWASMHash,
		&ledgerInt,
		&rel.TxHash,
		&rel.ObservedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReleaseNotFound
		}
		return nil, fmt.Errorf("get latest release: %w", err)
	}
	rel.Ledger = uint32(ledgerInt)
	return &rel, nil
}
