package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/hex"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/fleet"
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

func makeContractID(seed byte) string {
	var b [32]byte
	b[0] = seed
	res, _ := strkey.Encode(strkey.VersionByteContract, b[:])
	return res
}

func TestCLI_ExitCodesAndCommands(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	_, _ = db.Exec("TRUNCATE TABLE fleet_verifications, fleet_members, fleets CASCADE")

	// Set environment for CLI
	cliDBURL := os.Getenv("SFR_DATABASE_URL")
	if cliDBURL == "" {
		cliDBURL = "postgres://postgres@localhost:5433/sfr_test?sslmode=disable"
	}
	t.Setenv("SFR_DATABASE_URL", cliDBURL)

	// 1. Help -> ExitSuccess (0)
	var out, errOut bytes.Buffer
	code := run([]string{"--help"}, &out, &errOut)
	if code != ExitSuccess {
		t.Errorf("expected exit code %d for help, got %d", ExitSuccess, code)
	}

	// 2. Unknown command -> ExitInvalidInput (4)
	out.Reset()
	errOut.Reset()
	code = run([]string{"unknown_cmd"}, &out, &errOut)
	if code != ExitInvalidInput {
		t.Errorf("expected exit code %d for unknown command, got %d", ExitInvalidInput, code)
	}

	// 3. fleet list -> ExitSuccess (0)
	out.Reset()
	errOut.Reset()
	code = run([]string{"fleet", "list"}, &out, &errOut)
	if code != ExitSuccess {
		t.Errorf("expected exit code %d for fleet list, got %d (err: %s)", ExitSuccess, code, errOut.String())
	}

	// 4. fleet inspect without flags -> ExitInvalidInput (4)
	out.Reset()
	errOut.Reset()
	code = run([]string{"fleet", "inspect"}, &out, &errOut)
	if code != ExitInvalidInput {
		t.Errorf("expected exit code %d for missing flags, got %d", ExitInvalidInput, code)
	}

	// 5. Seed a fleet with 1 member for verify test
	fleetRepo := fleet.NewPostgresRepository(db)
	ownerID := makeContractID(90)
	tag := "vault-v1"
	fID := fleet.FleetID{Owner: ownerID, Tag: tag}
	wasmBytes := make([]byte, 32)
	wasmBytes[0] = 0xAA

	_ = fleetRepo.UpsertFleet(context.Background(), &fleet.Fleet{
		ID:                fID,
		CurrentWASMHash:   wasmBytes,
		FirstSeenLedger:   100,
		LastSeenLedger:    100,
		LastIndexedLedger: 100,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	})

	cid := makeContractID(91)
	_ = fleetRepo.UpsertMember(context.Background(), &fleet.FleetMember{
		ContractID:      cid,
		FleetID:         fID,
		WASMHash:        wasmBytes,
		FirstSeenLedger: 100,
		LastSeenLedger:  100,
		Active:          true,
	})

	// 6. fleet inspect with flags -> ExitSuccess (0)
	out.Reset()
	errOut.Reset()
	code = run([]string{"fleet", "inspect", "--owner", ownerID, "--tag", tag, "--json"}, &out, &errOut)
	if code != ExitSuccess {
		t.Errorf("expected exit code %d for fleet inspect, got %d (err: %s)", ExitSuccess, code, errOut.String())
	}

	// 7. fleet members with flags -> ExitSuccess (0)
	out.Reset()
	errOut.Reset()
	code = run([]string{"fleet", "members", "--owner", ownerID, "--tag", tag}, &out, &errOut)
	if code != ExitSuccess {
		t.Errorf("expected exit code %d for fleet members, got %d", ExitSuccess, code)
	}

	// 8. fleet verify with bad hex -> ExitInvalidInput (4)
	out.Reset()
	errOut.Reset()
	code = run([]string{"fleet", "verify", "--owner", ownerID, "--tag", tag, "--expected-wasm", "badhex"}, &out, &errOut)
	if code != ExitInvalidInput {
		t.Errorf("expected exit code %d for bad hex, got %d", ExitInvalidInput, code)
	}

	// 9. fleet verify with empty checkpoint (indexer lag / incomplete) -> ExitIncomplete (3)
	out.Reset()
	errOut.Reset()
	wasmHex := hex.EncodeToString(wasmBytes)
	code = run([]string{"fleet", "verify", "--owner", ownerID, "--tag", tag, "--expected-wasm", wasmHex}, &out, &errOut)
	// Because no checkpoint is indexed yet, the status is INCOMPLETE -> ExitIncomplete (3)
	if code != ExitIncomplete {
		t.Errorf("expected exit code %d for incomplete indexing, got %d", ExitIncomplete, code)
	}
}
