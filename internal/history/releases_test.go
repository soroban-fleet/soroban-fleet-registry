package history

import (
	"bytes"
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/soroban-fleet/soroban-fleet-registry/migrations"
	"github.com/stellar/go-stellar-sdk/strkey"
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

func makeContractID(seed byte) string {
	var b [32]byte
	b[0] = seed
	res, _ := strkey.Encode(strkey.VersionByteContract, b[:])
	return res
}

func TestReleasesRepository(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	_, _ = db.Exec("TRUNCATE TABLE fleet_releases CASCADE")

	repo := NewPostgresRepository(db)
	ctx := context.Background()

	ownerID := makeContractID(50)
	tag := "vault-v1"
	fID := FleetID{Owner: ownerID, Tag: tag}

	wasm1 := []byte{1, 1, 1}
	wasm2 := []byte{2, 2, 2}
	wasm3 := []byte{3, 3, 3}

	now := time.Now().UTC()

	// 1. Initial release (old=nil, new=wasm1)
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := repo.RecordReleaseTx(ctx, tx, ownerID, tag, nil, wasm1, 100, "tx_100", now); err != nil {
		_ = tx.Rollback()
		t.Fatalf("record release 1: %v", err)
	}
	_ = tx.Commit()

	// 2. Unchanged WASM hash should NOT insert a release
	tx2, _ := db.BeginTx(ctx, nil)
	_ = repo.RecordReleaseTx(ctx, tx2, ownerID, tag, wasm1, wasm1, 110, "tx_110", now)
	_ = tx2.Commit()

	// 3. Second release (old=wasm1, new=wasm2)
	tx3, _ := db.BeginTx(ctx, nil)
	_ = repo.RecordReleaseTx(ctx, tx3, ownerID, tag, wasm1, wasm2, 200, "tx_200", now.Add(time.Hour))
	_ = tx3.Commit()

	// 4. Duplicate release attempt (idempotency test for replay of tx_200 at ledger 200)
	txDup, _ := db.BeginTx(ctx, nil)
	_ = repo.RecordReleaseTx(ctx, txDup, ownerID, tag, wasm1, wasm2, 200, "tx_200", now.Add(time.Hour))
	_ = txDup.Commit()

	// 5. Third release (old=wasm2, new=wasm3)
	tx4, _ := db.BeginTx(ctx, nil)
	_ = repo.RecordReleaseTx(ctx, tx4, ownerID, tag, wasm2, wasm3, 300, "tx_300", now.Add(2*time.Hour))
	_ = tx4.Commit()

	// Verify ListReleases
	releases, total, err := repo.ListReleases(ctx, fID, 10, 0)
	if err != nil {
		t.Fatalf("list releases: %v", err)
	}
	if total != 3 || len(releases) != 3 {
		t.Fatalf("expected 3 total releases, got total=%d len=%d", total, len(releases))
	}

	// Must be ordered by ledger DESC: 300, 200, 100
	if releases[0].Ledger != 300 || releases[1].Ledger != 200 || releases[2].Ledger != 100 {
		t.Errorf("releases not ordered by ledger DESC: got %d, %d, %d", releases[0].Ledger, releases[1].Ledger, releases[2].Ledger)
	}

	// Verify LatestRelease
	latest, err := repo.GetLatestRelease(ctx, fID)
	if err != nil {
		t.Fatalf("get latest release: %v", err)
	}
	if latest.Ledger != 300 || !bytes.Equal(latest.NewWASMHash, wasm3) {
		t.Errorf("latest release mismatch: got %+v", latest)
	}
}
