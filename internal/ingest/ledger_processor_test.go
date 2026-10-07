package ingest

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/cap85"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/fleet"
	"github.com/soroban-fleet/soroban-fleet-registry/migrations"
	"github.com/stellar/go-stellar-sdk/xdr"
)

func getTestDB(t *testing.T) *sql.DB {
	dbURL := os.Getenv("SFR_DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres@localhost:5433/sfr_test?sslmode=disable"
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Skipf("cannot connect to db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("db not reachable: %v", err)
	}
	if err := migrations.Up(db); err != nil {
		t.Fatalf("failed to apply migrations: %v", err)
	}
	return db
}

func TestLedgerProcessor_ProcessLedgerTx(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	_, _ = db.Exec("TRUNCATE TABLE fleet_members, fleets CASCADE")

	repo := fleet.NewPostgresRepository(db)
	mm := fleet.NewMembershipManager(repo)
	proc := NewLedgerProcessor(repo, mm, nil, nil)
	ctx := context.Background()

	ownerID := makeContractID(10)
	cid := makeContractID(11)
	tag := "vault-v1"
	wasmHash := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}

	ownerScAddr, _ := cap85.AddressToScAddress(ownerID)
	extRef := xdr.ContractExecutableExternalRef{
		ExecutableOwner: ownerScAddr,
		Tag:             xdr.ScString(tag),
	}
	newExec, _ := xdr.NewContractExecutable(xdr.ContractExecutableTypeContractExecutableExternalRef, extRef)

	// Step 1: Ledger 100 has tag change and contract created
	changes := LedgerChanges{
		Sequence:      100,
		CloseTimeUnix: 1700000000,
		TagChanges: []TagExecutableChange{
			{
				OwnerAddress: ownerID,
				Tag:          tag,
				NewWASMHash:  wasmHash,
				TxHash:       "tx1",
			},
		},
		ContractChanges: []ContractInstanceChange{
			{
				ContractID: cid,
				NewExec:    &newExec,
				TxHash:     "tx2",
			},
		},
	}

	tx, err := repo.BeginTx(ctx)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	stats, err := proc.ProcessLedgerTx(ctx, tx, changes)
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("process ledger: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}

	if stats.ChangesProcessed != 2 {
		t.Errorf("expected 2 changes processed, got %d", stats.ChangesProcessed)
	}
	if stats.FleetsUpdated != 2 { // 1 from tag change + 1 from member fleet upsert
		t.Errorf("expected 2 fleet updates, got %d", stats.FleetsUpdated)
	}
	if stats.MembersUpdated != 1 {
		t.Errorf("expected 1 member update, got %d", stats.MembersUpdated)
	}

	// Verify fleet in DB
	fID := fleet.FleetID{Owner: ownerID, Tag: tag}
	f, err := repo.GetFleet(ctx, fID)
	if err != nil {
		t.Fatalf("get fleet: %v", err)
	}
	if !bytes.Equal(f.CurrentWASMHash, wasmHash) {
		t.Errorf("wasm hash mismatch")
	}
	if f.MemberCount != 1 {
		t.Errorf("expected member count 1, got %d", f.MemberCount)
	}

	// Step 2: Ledger 200 deactivates member (exec changes to direct wasm)
	directExec, _ := xdr.NewContractExecutable(xdr.ContractExecutableTypeContractExecutableWasm, xdr.Hash([32]byte{9}))
	changes2 := LedgerChanges{
		Sequence:      200,
		CloseTimeUnix: 1700001000,
		ContractChanges: []ContractInstanceChange{
			{
				ContractID: cid,
				OldExec:    &newExec,
				NewExec:    &directExec,
				TxHash:     "tx3",
			},
		},
	}

	tx2, _ := repo.BeginTx(ctx)
	stats2, err := proc.ProcessLedgerTx(ctx, tx2, changes2)
	if err != nil {
		_ = tx2.Rollback()
		t.Fatalf("process ledger 2: %v", err)
	}
	_ = tx2.Commit()

	if stats2.MembersUpdated != 1 {
		t.Errorf("expected 1 member update in deactivation, got %d", stats2.MembersUpdated)
	}

	f, _ = repo.GetFleet(ctx, fID)
	if f.MemberCount != 0 {
		t.Errorf("expected fleet member count 0 after deactivation, got %d", f.MemberCount)
	}
}
