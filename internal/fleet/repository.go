package fleet

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// Repository defines data access methods for fleets and fleet members.
type Repository interface {
	BeginTx(ctx context.Context) (*sql.Tx, error)

	GetFleet(ctx context.Context, id FleetID) (*Fleet, error)
	GetFleetTx(ctx context.Context, tx *sql.Tx, id FleetID) (*Fleet, error)
	ListFleets(ctx context.Context, limit, offset int) ([]Fleet, int64, error)
	UpsertFleet(ctx context.Context, fleet *Fleet) error
	UpsertFleetTx(ctx context.Context, tx *sql.Tx, fleet *Fleet) error

	GetMember(ctx context.Context, contractID string) (*FleetMember, error)
	ListMembers(ctx context.Context, fleetID FleetID, activeOnly bool, limit, offset int) ([]FleetMember, int64, error)
	UpsertMember(ctx context.Context, member *FleetMember) error
	UpsertMemberTx(ctx context.Context, tx *sql.Tx, member *FleetMember) error
	GetActiveMembers(ctx context.Context, fleetID FleetID) ([]FleetMember, error)
	SetMemberActiveTx(ctx context.Context, tx *sql.Tx, contractID string, active bool, lastSeenLedger uint32) error
	RefreshFleetMemberCountTx(ctx context.Context, tx *sql.Tx, fleetID FleetID) error
}

// PostgresRepository implements Repository using PostgreSQL.
type PostgresRepository struct {
	db *sql.DB
}

// NewPostgresRepository creates a new PostgresRepository.
func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// BeginTx starts a database transaction.
func (r *PostgresRepository) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return r.db.BeginTx(ctx, nil)
}

// GetFleet fetches a fleet by its (owner, tag) identity.
func (r *PostgresRepository) GetFleet(ctx context.Context, id FleetID) (*Fleet, error) {
	query := `
		SELECT owner_address, tag, current_wasm_hash, member_count,
		       first_seen_ledger, last_seen_ledger, last_indexed_ledger,
		       created_at, updated_at
		FROM fleets
		WHERE owner_address = $1 AND tag = $2
	`
	row := r.db.QueryRowContext(ctx, query, id.Owner, id.Tag)

	var f Fleet
	var firstSeen, lastSeen, lastIndexed int64
	err := row.Scan(
		&f.ID.Owner,
		&f.ID.Tag,
		&f.CurrentWASMHash,
		&f.MemberCount,
		&firstSeen,
		&lastSeen,
		&lastIndexed,
		&f.CreatedAt,
		&f.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrFleetNotFound
		}
		return nil, fmt.Errorf("get fleet %v: %w", id, err)
	}

	f.FirstSeenLedger = uint32(firstSeen)
	f.LastSeenLedger = uint32(lastSeen)
	f.LastIndexedLedger = uint32(lastIndexed)

	return &f, nil
}

// GetFleetTx fetches a fleet within an existing transaction.
func (r *PostgresRepository) GetFleetTx(ctx context.Context, tx *sql.Tx, id FleetID) (*Fleet, error) {
	query := `
		SELECT owner_address, tag, current_wasm_hash, member_count,
		       first_seen_ledger, last_seen_ledger, last_indexed_ledger,
		       created_at, updated_at
		FROM fleets
		WHERE owner_address = $1 AND tag = $2
	`
	row := tx.QueryRowContext(ctx, query, id.Owner, id.Tag)

	var f Fleet
	var firstSeen, lastSeen, lastIndexed int64
	err := row.Scan(
		&f.ID.Owner,
		&f.ID.Tag,
		&f.CurrentWASMHash,
		&f.MemberCount,
		&firstSeen,
		&lastSeen,
		&lastIndexed,
		&f.CreatedAt,
		&f.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrFleetNotFound
		}
		return nil, fmt.Errorf("get fleet tx %v: %w", id, err)
	}

	f.FirstSeenLedger = uint32(firstSeen)
	f.LastSeenLedger = uint32(lastSeen)
	f.LastIndexedLedger = uint32(lastIndexed)

	return &f, nil
}

// ListFleets returns a paginated list of fleets and the total count.
func (r *PostgresRepository) ListFleets(ctx context.Context, limit, offset int) ([]Fleet, int64, error) {
	var total int64
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM fleets").Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count fleets: %w", err)
	}

	query := `
		SELECT owner_address, tag, current_wasm_hash, member_count,
		       first_seen_ledger, last_seen_ledger, last_indexed_ledger,
		       created_at, updated_at
		FROM fleets
		ORDER BY created_at DESC, owner_address ASC, tag ASC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list fleets: %w", err)
	}
	defer rows.Close()

	var fleets []Fleet
	for rows.Next() {
		var f Fleet
		var firstSeen, lastSeen, lastIndexed int64
		if err := rows.Scan(
			&f.ID.Owner,
			&f.ID.Tag,
			&f.CurrentWASMHash,
			&f.MemberCount,
			&firstSeen,
			&lastSeen,
			&lastIndexed,
			&f.CreatedAt,
			&f.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("scan fleet row: %w", err)
		}
		f.FirstSeenLedger = uint32(firstSeen)
		f.LastSeenLedger = uint32(lastSeen)
		f.LastIndexedLedger = uint32(lastIndexed)
		fleets = append(fleets, f)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return fleets, total, nil
}

// UpsertFleet inserts or updates a fleet row.
func (r *PostgresRepository) UpsertFleet(ctx context.Context, fleet *Fleet) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := r.UpsertFleetTx(ctx, tx, fleet); err != nil {
		return err
	}
	return tx.Commit()
}

// UpsertFleetTx inserts or updates a fleet within an existing transaction.
func (r *PostgresRepository) UpsertFleetTx(ctx context.Context, tx *sql.Tx, fleet *Fleet) error {
	now := time.Now().UTC()
	if fleet.CreatedAt.IsZero() {
		fleet.CreatedAt = now
	}
	fleet.UpdatedAt = now

	query := `
		INSERT INTO fleets (
			owner_address, tag, current_wasm_hash, member_count,
			first_seen_ledger, last_seen_ledger, last_indexed_ledger,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (owner_address, tag) DO UPDATE SET
			current_wasm_hash = CASE 
				WHEN EXCLUDED.current_wasm_hash IS NOT NULL AND length(EXCLUDED.current_wasm_hash) > 0 
				THEN EXCLUDED.current_wasm_hash 
				ELSE fleets.current_wasm_hash 
			END,
			member_count = EXCLUDED.member_count,
			first_seen_ledger = LEAST(fleets.first_seen_ledger, EXCLUDED.first_seen_ledger),
			last_seen_ledger = GREATEST(fleets.last_seen_ledger, EXCLUDED.last_seen_ledger),
			last_indexed_ledger = GREATEST(fleets.last_indexed_ledger, EXCLUDED.last_indexed_ledger),
			updated_at = EXCLUDED.updated_at
	`
	_, err := tx.ExecContext(
		ctx, query,
		fleet.ID.Owner,
		fleet.ID.Tag,
		fleet.CurrentWASMHash,
		fleet.MemberCount,
		fleet.FirstSeenLedger,
		fleet.LastSeenLedger,
		fleet.LastIndexedLedger,
		fleet.CreatedAt,
		fleet.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("upsert fleet %v: %w", fleet.ID, err)
	}
	return nil
}

// GetMember retrieves a fleet member by contract ID.
func (r *PostgresRepository) GetMember(ctx context.Context, contractID string) (*FleetMember, error) {
	query := `
		SELECT contract_id, owner_address, tag, wasm_hash,
		       first_seen_ledger, last_seen_ledger, last_verified_ledger, active
		FROM fleet_members
		WHERE contract_id = $1
	`
	row := r.db.QueryRowContext(ctx, query, contractID)

	var m FleetMember
	var firstSeen, lastSeen int64
	var lastVerified sql.NullInt64

	err := row.Scan(
		&m.ContractID,
		&m.FleetID.Owner,
		&m.FleetID.Tag,
		&m.WASMHash,
		&firstSeen,
		&lastSeen,
		&lastVerified,
		&m.Active,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrFleetMemberNotFound
		}
		return nil, fmt.Errorf("get member %s: %w", contractID, err)
	}

	m.FirstSeenLedger = uint32(firstSeen)
	m.LastSeenLedger = uint32(lastSeen)
	if lastVerified.Valid {
		v := uint32(lastVerified.Int64)
		m.LastVerifiedLedger = &v
	}

	return &m, nil
}

// ListMembers returns paginated members of a fleet.
func (r *PostgresRepository) ListMembers(ctx context.Context, fleetID FleetID, activeOnly bool, limit, offset int) ([]FleetMember, int64, error) {
	countQuery := "SELECT COUNT(*) FROM fleet_members WHERE owner_address = $1 AND tag = $2"
	if activeOnly {
		countQuery += " AND active = TRUE"
	}

	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, fleetID.Owner, fleetID.Tag).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count fleet members: %w", err)
	}

	query := `
		SELECT contract_id, owner_address, tag, wasm_hash,
		       first_seen_ledger, last_seen_ledger, last_verified_ledger, active
		FROM fleet_members
		WHERE owner_address = $1 AND tag = $2
	`
	if activeOnly {
		query += " AND active = TRUE"
	}
	query += " ORDER BY active DESC, first_seen_ledger ASC, contract_id ASC LIMIT $3 OFFSET $4"

	rows, err := r.db.QueryContext(ctx, query, fleetID.Owner, fleetID.Tag, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list members: %w", err)
	}
	defer rows.Close()

	var members []FleetMember
	for rows.Next() {
		var m FleetMember
		var firstSeen, lastSeen int64
		var lastVerified sql.NullInt64

		if err := rows.Scan(
			&m.ContractID,
			&m.FleetID.Owner,
			&m.FleetID.Tag,
			&m.WASMHash,
			&firstSeen,
			&lastSeen,
			&lastVerified,
			&m.Active,
		); err != nil {
			return nil, 0, fmt.Errorf("scan member row: %w", err)
		}
		m.FirstSeenLedger = uint32(firstSeen)
		m.LastSeenLedger = uint32(lastSeen)
		if lastVerified.Valid {
			v := uint32(lastVerified.Int64)
			m.LastVerifiedLedger = &v
		}
		members = append(members, m)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	return members, total, nil
}

// UpsertMember inserts or updates a member.
func (r *PostgresRepository) UpsertMember(ctx context.Context, member *FleetMember) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := r.UpsertMemberTx(ctx, tx, member); err != nil {
		return err
	}
	if err := r.RefreshFleetMemberCountTx(ctx, tx, member.FleetID); err != nil {
		return err
	}
	return tx.Commit()
}

// UpsertMemberTx inserts or updates a fleet member within an active transaction.
func (r *PostgresRepository) UpsertMemberTx(ctx context.Context, tx *sql.Tx, member *FleetMember) error {
	query := `
		INSERT INTO fleet_members (
			contract_id, owner_address, tag, wasm_hash,
			first_seen_ledger, last_seen_ledger, last_verified_ledger, active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (contract_id) DO UPDATE SET
			owner_address = EXCLUDED.owner_address,
			tag = EXCLUDED.tag,
			wasm_hash = EXCLUDED.wasm_hash,
			first_seen_ledger = LEAST(fleet_members.first_seen_ledger, EXCLUDED.first_seen_ledger),
			last_seen_ledger = GREATEST(fleet_members.last_seen_ledger, EXCLUDED.last_seen_ledger),
			last_verified_ledger = COALESCE(EXCLUDED.last_verified_ledger, fleet_members.last_verified_ledger),
			active = EXCLUDED.active
	`
	var lastVerified *int64
	if member.LastVerifiedLedger != nil {
		val := int64(*member.LastVerifiedLedger)
		lastVerified = &val
	}

	_, err := tx.ExecContext(
		ctx, query,
		member.ContractID,
		member.FleetID.Owner,
		member.FleetID.Tag,
		member.WASMHash,
		member.FirstSeenLedger,
		member.LastSeenLedger,
		lastVerified,
		member.Active,
	)
	if err != nil {
		return fmt.Errorf("upsert member %s: %w", member.ContractID, err)
	}
	return nil
}

// GetActiveMembers retrieves all active members of a fleet for verification.
func (r *PostgresRepository) GetActiveMembers(ctx context.Context, fleetID FleetID) ([]FleetMember, error) {
	query := `
		SELECT contract_id, owner_address, tag, wasm_hash,
		       first_seen_ledger, last_seen_ledger, last_verified_ledger, active
		FROM fleet_members
		WHERE owner_address = $1 AND tag = $2 AND active = TRUE
		ORDER BY contract_id ASC
	`
	rows, err := r.db.QueryContext(ctx, query, fleetID.Owner, fleetID.Tag)
	if err != nil {
		return nil, fmt.Errorf("get active members: %w", err)
	}
	defer rows.Close()

	var members []FleetMember
	for rows.Next() {
		var m FleetMember
		var firstSeen, lastSeen int64
		var lastVerified sql.NullInt64

		if err := rows.Scan(
			&m.ContractID,
			&m.FleetID.Owner,
			&m.FleetID.Tag,
			&m.WASMHash,
			&firstSeen,
			&lastSeen,
			&lastVerified,
			&m.Active,
		); err != nil {
			return nil, fmt.Errorf("scan active member: %w", err)
		}
		m.FirstSeenLedger = uint32(firstSeen)
		m.LastSeenLedger = uint32(lastSeen)
		if lastVerified.Valid {
			v := uint32(lastVerified.Int64)
			m.LastVerifiedLedger = &v
		}
		members = append(members, m)
	}

	return members, rows.Err()
}

// SetMemberActiveTx updates a member's active status.
func (r *PostgresRepository) SetMemberActiveTx(ctx context.Context, tx *sql.Tx, contractID string, active bool, lastSeenLedger uint32) error {
	query := `
		UPDATE fleet_members
		SET active = $1, last_seen_ledger = GREATEST(last_seen_ledger, $2)
		WHERE contract_id = $3
	`
	res, err := tx.ExecContext(ctx, query, active, lastSeenLedger, contractID)
	if err != nil {
		return fmt.Errorf("update member %s active=%v: %w", contractID, active, err)
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return ErrFleetMemberNotFound
	}
	return nil
}

// RefreshFleetMemberCountTx updates member_count on the fleets table to match active members.
func (r *PostgresRepository) RefreshFleetMemberCountTx(ctx context.Context, tx *sql.Tx, fleetID FleetID) error {
	query := `
		UPDATE fleets
		SET member_count = (
			SELECT COUNT(*) FROM fleet_members
			WHERE owner_address = $1 AND tag = $2 AND active = TRUE
		), updated_at = NOW()
		WHERE owner_address = $1 AND tag = $2
	`
	_, err := tx.ExecContext(ctx, query, fleetID.Owner, fleetID.Tag)
	if err != nil {
		return fmt.Errorf("refresh member count for fleet %v: %w", fleetID, err)
	}
	return nil
}
