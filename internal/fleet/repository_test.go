package fleet

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/soroban-fleet/soroban-fleet-registry/migrations"
)

func getTestDB(t *testing.T) *sql.DB {
	dbURL := os.Getenv("SFR_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres@localhost:5433/sfr_test?sslmode=disable"
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Skipf("cannot connect to test db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("test db not reachable: %v", err)
	}
	if err := migrations.Up(db); err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}
	return db
}

func TestFleetRepository_CRUD(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	repo := NewPostgresRepository(db)
	ctx := context.Background()

	fleetID := FleetID{
		Owner: "CAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		Tag:   "vault-v1",
	}

	// 1. Not found initially
	_, err := repo.GetFleet(ctx, fleetID)
	if !errors.Is(err, ErrFleetNotFound) {
		t.Fatalf("expected ErrFleetNotFound, got %v", err)
	}

	// 2. Upsert Fleet
	wasmHash := []byte("wasm_hash_1234567890123456789012")
	f := &Fleet{
		ID:                fleetID,
		CurrentWASMHash:   wasmHash,
		MemberCount:       0,
		FirstSeenLedger:   100,
		LastSeenLedger:    100,
		LastIndexedLedger: 100,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	if err := repo.UpsertFleet(ctx, f); err != nil {
		t.Fatalf("failed to upsert fleet: %v", err)
	}

	// 3. Get Fleet
	fetched, err := repo.GetFleet(ctx, fleetID)
	if err != nil {
		t.Fatalf("failed to get fleet: %v", err)
	}
	if fetched.ID != fleetID {
		t.Errorf("fleet ID mismatch: %+v", fetched.ID)
	}
	if !bytes.Equal(fetched.CurrentWASMHash, wasmHash) {
		t.Errorf("wasm hash mismatch")
	}

	// 4. Upsert Member
	member1 := &FleetMember{
		ContractID:      "CB111111111111111111111111111111111111111111111111111111",
		FleetID:         fleetID,
		WASMHash:        wasmHash,
		FirstSeenLedger: 100,
		LastSeenLedger:  100,
		Active:          true,
	}
	if err := repo.UpsertMember(ctx, member1); err != nil {
		t.Fatalf("failed to upsert member: %v", err)
	}

	// Verify member count updated
	fetched, _ = repo.GetFleet(ctx, fleetID)
	if fetched.MemberCount != 1 {
		t.Errorf("expected member count 1, got %d", fetched.MemberCount)
	}

	// 5. Get Member
	m, err := repo.GetMember(ctx, member1.ContractID)
	if err != nil {
		t.Fatalf("failed to get member: %v", err)
	}
	if m.ContractID != member1.ContractID || !m.Active {
		t.Errorf("member mismatch: %+v", m)
	}

	// 6. List Members
	members, total, err := repo.ListMembers(ctx, fleetID, true, 10, 0)
	if err != nil {
		t.Fatalf("failed to list members: %v", err)
	}
	if total != 1 || len(members) != 1 {
		t.Errorf("expected 1 member, got total=%d len=%d", total, len(members))
	}

	// 7. Deactivate Member
	tx, err := repo.BeginTx(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := repo.SetMemberActiveTx(ctx, tx, member1.ContractID, false, 150); err != nil {
		_ = tx.Rollback()
		t.Fatalf("deactivate member: %v", err)
	}
	if err := repo.RefreshFleetMemberCountTx(ctx, tx, fleetID); err != nil {
		_ = tx.Rollback()
		t.Fatalf("refresh count: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	// Active list should be 0, all list should be 1
	activeMembers, activeTotal, err := repo.ListMembers(ctx, fleetID, true, 10, 0)
	if err != nil || activeTotal != 0 || len(activeMembers) != 0 {
		t.Errorf("expected 0 active members, got total=%d len=%d", activeTotal, len(activeMembers))
	}

	allMembers, allTotal, err := repo.ListMembers(ctx, fleetID, false, 10, 0)
	if err != nil || allTotal != 1 || len(allMembers) != 1 {
		t.Errorf("expected 1 total member in historical set, got total=%d len=%d", allTotal, len(allMembers))
	}

	// Fleet member_count should be 0
	fetched, _ = repo.GetFleet(ctx, fleetID)
	if fetched.MemberCount != 0 {
		t.Errorf("expected fleet member_count 0 after deactivation, got %d", fetched.MemberCount)
	}
}
