package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/hex"
	"os"
	"testing"

	_ "github.com/lib/pq"
	"github.com/soroban-fleet/soroban-fleet-registry/fixtures"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/cap85"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/fleet"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/history"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/ingest"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/rpc"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/verification"
	"github.com/soroban-fleet/soroban-fleet-registry/migrations"
	protocol "github.com/stellar/go-stellar-sdk/protocols/rpc"
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

type controllableResolver struct {
	overrides map[string]cap85.ResolvedExecutable
	fallback  cap85.ResolvedExecutable
}

func (r *controllableResolver) ResolveContractExecutable(ctx context.Context, contractID string) (cap85.ResolvedExecutable, error) {
	if res, ok := r.overrides[contractID]; ok {
		return res, nil
	}
	return r.fallback, nil
}

func (r *controllableResolver) ResolveContractExecutablesBatch(ctx context.Context, contractIDs []string) (map[string]cap85.ResolvedExecutable, error) {
	res := make(map[string]cap85.ResolvedExecutable, len(contractIDs))
	for _, cid := range contractIDs {
		if o, ok := r.overrides[cid]; ok {
			res[cid] = o
		} else {
			res[cid] = r.fallback
		}
	}
	return res, nil
}

type testRPC struct {
	currentLedger uint32
}

func (m *testRPC) GetLatestLedger(ctx context.Context) (rpc.LatestLedger, error) {
	return rpc.LatestLedger{Sequence: m.currentLedger}, nil
}
func (m *testRPC) GetLedgerEntries(ctx context.Context, keys []xdr.LedgerKey) (rpc.LedgerEntriesResponse, error) {
	return rpc.LedgerEntriesResponse{}, nil
}
func (m *testRPC) GetLedgers(ctx context.Context, startLedger uint32, limit uint32) (protocol.GetLedgersResponse, error) {
	return protocol.GetLedgersResponse{}, nil
}

func TestFlagshipEndToEndScenario(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	_, _ = db.Exec("TRUNCATE TABLE fleet_verifications, fleet_releases, fleet_members, fleets, indexer_checkpoints CASCADE")

	ctx := context.Background()
	healthyFixture, err := fixtures.LoadHealthyFixture()
	if err != nil {
		t.Fatalf("load healthy fixture: %v", err)
	}

	ownerID := healthyFixture.OwnerAddress
	tag := healthyFixture.Tag
	fleetID := fleet.FleetID{Owner: ownerID, Tag: tag}

	wasmBytesA, _ := hex.DecodeString(healthyFixture.ExpectedWASM)
	contractA := healthyFixture.Members[0].ContractID
	contractB := healthyFixture.Members[1].ContractID
	contractC := healthyFixture.Members[2].ContractID

	// Setup database repositories and services
	fleetRepo := fleet.NewPostgresRepository(db)
	membership := fleet.NewMembershipManager(fleetRepo)
	checkpoints := ingest.NewPostgresCheckpointStore(db)
	releases := history.NewPostgresRepository(db)

	mockRPCClient := &testRPC{currentLedger: 1000}
	resolver := &controllableResolver{
		overrides: make(map[string]cap85.ResolvedExecutable),
		fallback: cap85.ResolvedExecutable{
			Kind:     "EXTERNAL_REF",
			Fleet:    &fleetID,
			WASMHash: wasmBytesA,
		},
	}

	processor := ingest.NewLedgerProcessor(fleetRepo, membership, resolver, releases)
	verifier := verification.NewEngine(db, fleetRepo, checkpoints, resolver, mockRPCClient)

	// =========================================================================
	// PHASE 1: Initial deployment (Ledger 1000)
	// Fleet: Contract A, Contract B, Contract C all point to WASM A
	// =========================================================================
	ownerScAddr, _ := cap85.AddressToScAddress(ownerID)
	extRef := xdr.ContractExecutableExternalRef{
		ExecutableOwner: ownerScAddr,
		Tag:             xdr.ScString(tag),
	}
	exec, _ := xdr.NewContractExecutable(xdr.ContractExecutableTypeContractExecutableExternalRef, extRef)

	ledger1000Changes := ingest.LedgerChanges{
		Sequence:      1000,
		CloseTimeUnix: 1700000000,
		TagChanges: []ingest.TagExecutableChange{
			{
				OwnerAddress: ownerID,
				Tag:          tag,
				OldWASMHash:  nil,
				NewWASMHash:  wasmBytesA,
				TxHash:       "tx_init",
			},
		},
		ContractChanges: []ingest.ContractInstanceChange{
			{ContractID: contractA, NewExec: &exec, TxHash: "tx_ca"},
			{ContractID: contractB, NewExec: &exec, TxHash: "tx_cb"},
			{ContractID: contractC, NewExec: &exec, TxHash: "tx_cc"},
		},
	}

	tx1, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx1: %v", err)
	}
	stats1, err := processor.ProcessLedgerTx(ctx, tx1, ledger1000Changes)
	if err != nil {
		_ = tx1.Rollback()
		t.Fatalf("process ledger 1000: %v", err)
	}
	_ = checkpoints.SetCheckpointTx(ctx, tx1, ingest.DefaultStreamName, 1000)
	if err := tx1.Commit(); err != nil {
		t.Fatalf("commit tx1: %v", err)
	}

	if stats1.MembersUpdated != 3 {
		t.Errorf("expected 3 members updated in ledger 1000, got %d", stats1.MembersUpdated)
	}

	// VERIFY PHASE 1: Expected HEALTHY
	res1, err := verifier.VerifyFleet(ctx, fleetID, wasmBytesA)
	if err != nil {
		t.Fatalf("verification 1 failed: %v", err)
	}
	if res1.Status != verification.StatusHealthy {
		t.Fatalf("expected HEALTHY in phase 1, got %s", res1.Status)
	}
	if res1.TotalMembers != 3 || res1.MatchingMembers != 3 || res1.MismatchingMembers != 0 || res1.MissingMembers != 0 {
		t.Errorf("phase 1 count mismatch: %+v", res1)
	}

	// =========================================================================
	// PHASE 2: Fleet Upgrade to WASM B (Ledger 1050)
	// Owner updates tag to WASM B. All 3 contracts now resolve to WASM B.
	// =========================================================================
	releaseFixture, err := fixtures.LoadReleaseFixture()
	if err != nil {
		t.Fatalf("load release fixture: %v", err)
	}
	wasmBytesB, _ := hex.DecodeString(releaseFixture.NewWASM)

	mockRPCClient.currentLedger = 1050
	resolver.fallback = cap85.ResolvedExecutable{
		Kind:     "EXTERNAL_REF",
		Fleet:    &fleetID,
		WASMHash: wasmBytesB,
	}

	ledger1050Changes := ingest.LedgerChanges{
		Sequence:      1050,
		CloseTimeUnix: 1700003000,
		TagChanges: []ingest.TagExecutableChange{
			{
				OwnerAddress: ownerID,
				Tag:          tag,
				OldWASMHash:  wasmBytesA,
				NewWASMHash:  wasmBytesB,
				TxHash:       releaseFixture.TxHash,
			},
		},
	}

	tx2, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx2: %v", err)
	}
	stats2, err := processor.ProcessLedgerTx(ctx, tx2, ledger1050Changes)
	if err != nil {
		_ = tx2.Rollback()
		t.Fatalf("process ledger 1050: %v", err)
	}
	_ = checkpoints.SetCheckpointTx(ctx, tx2, ingest.DefaultStreamName, 1050)
	if err := tx2.Commit(); err != nil {
		t.Fatalf("commit tx2: %v", err)
	}

	if stats2.ReleasesDetected != 1 {
		t.Errorf("expected 1 release detected in ledger 1050, got %d", stats2.ReleasesDetected)
	}

	// Verify release was recorded in history
	latestRel, err := releases.GetLatestRelease(ctx, fleetID)
	if err != nil {
		t.Fatalf("get latest release: %v", err)
	}
	if !bytes.Equal(latestRel.NewWASMHash, wasmBytesB) || latestRel.Ledger != 1050 {
		t.Errorf("unexpected latest release: %+v", latestRel)
	}

	// VERIFY PHASE 2: Expected HEALTHY against WASM B
	res2, err := verifier.VerifyFleet(ctx, fleetID, wasmBytesB)
	if err != nil {
		t.Fatalf("verification 2 failed: %v", err)
	}
	if res2.Status != verification.StatusHealthy {
		t.Fatalf("expected HEALTHY in phase 2, got %s", res2.Status)
	}
	if res2.TotalMembers != 3 || res2.MatchingMembers != 3 || res2.MismatchingMembers != 0 || res2.MissingMembers != 0 {
		t.Errorf("phase 2 count mismatch: %+v", res2)
	}

	// =========================================================================
	// PHASE 3: DRIFT SIMULATION
	// Simulate one stale member (Contract C still resolving to WASM A)
	// =========================================================================
	resolver.overrides[contractC] = cap85.ResolvedExecutable{
		Kind:     "EXTERNAL_REF",
		Fleet:    &fleetID,
		WASMHash: wasmBytesA, // Stale!
	}

	res3, err := verifier.VerifyFleet(ctx, fleetID, wasmBytesB)
	if err != nil {
		t.Fatalf("verification 3 failed: %v", err)
	}

	// Critical correctness assertion from Section 27:
	// total = 3, matching = 2, mismatching = 1, missing = 0, status = DRIFT
	if res3.Status != verification.StatusDrift {
		t.Fatalf("expected DRIFT in phase 3, got %s", res3.Status)
	}
	if res3.TotalMembers != 3 {
		t.Errorf("expected TotalMembers = 3, got %d", res3.TotalMembers)
	}
	if res3.MatchingMembers != 2 {
		t.Errorf("expected MatchingMembers = 2, got %d", res3.MatchingMembers)
	}
	if res3.MismatchingMembers != 1 {
		t.Errorf("expected MismatchingMembers = 1, got %d", res3.MismatchingMembers)
	}
	if res3.MissingMembers != 0 {
		t.Errorf("expected MissingMembers = 0, got %d", res3.MissingMembers)
	}
}
