package verification

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/cap85"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/fleet"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/rpc"
	"github.com/soroban-fleet/soroban-fleet-registry/migrations"
	protocol "github.com/stellar/go-stellar-sdk/protocols/rpc"
	"github.com/stellar/go-stellar-sdk/strkey"
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

func makeContractID(seed byte) string {
	var b [32]byte
	b[0] = seed
	res, _ := strkey.Encode(strkey.VersionByteContract, b[:])
	return res
}

type mockCheckpoints struct {
	cp uint32
}

func (m *mockCheckpoints) GetCheckpoint(ctx context.Context, stream string) (uint32, error) {
	return m.cp, nil
}
func (m *mockCheckpoints) SetCheckpointTx(ctx context.Context, tx *sql.Tx, stream string, ledger uint32) error {
	m.cp = ledger
	return nil
}

type mockRPC struct {
	latestLedger uint32
}

func (m *mockRPC) GetLatestLedger(ctx context.Context) (rpc.LatestLedger, error) {
	return rpc.LatestLedger{Sequence: m.latestLedger}, nil
}
func (m *mockRPC) GetLedgerEntries(ctx context.Context, keys []xdr.LedgerKey) (rpc.LedgerEntriesResponse, error) {
	return rpc.LedgerEntriesResponse{}, nil
}
func (m *mockRPC) GetLedgers(ctx context.Context, startLedger uint32, limit uint32) (protocol.GetLedgersResponse, error) {
	return protocol.GetLedgersResponse{}, nil
}

type mockResolver struct {
	resolved map[string]cap85.ResolvedExecutable
	broken   bool
}

func (m *mockResolver) ResolveContractExecutable(ctx context.Context, contractID string) (cap85.ResolvedExecutable, error) {
	if m.broken {
		return cap85.ResolvedExecutable{}, cap85.ErrBrokenReference
	}
	if res, ok := m.resolved[contractID]; ok {
		return res, nil
	}
	return cap85.ResolvedExecutable{}, cap85.ErrBrokenReference
}

func (m *mockResolver) ResolveContractExecutablesBatch(ctx context.Context, contractIDs []string) (map[string]cap85.ResolvedExecutable, error) {
	if m.broken {
		return nil, cap85.ErrBrokenReference
	}
	res := make(map[string]cap85.ResolvedExecutable)
	for _, cid := range contractIDs {
		if r, ok := m.resolved[cid]; ok {
			res[cid] = r
		}
	}
	return res, nil
}

func TestVerifier_Healthy(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	_, _ = db.Exec("TRUNCATE TABLE fleet_verifications, fleet_members, fleets CASCADE")

	repo := fleet.NewPostgresRepository(db)
	ownerID := makeContractID(1)
	fleetID := FleetID{Owner: ownerID, Tag: "vault-v1"}
	wasm := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}

	_ = repo.UpsertFleet(context.Background(), &fleet.Fleet{
		ID:                fleetID,
		CurrentWASMHash:   wasm,
		FirstSeenLedger:   100,
		LastSeenLedger:    100,
		LastIndexedLedger: 100,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	})

	cid1 := makeContractID(10)
	cid2 := makeContractID(11)
	_ = repo.UpsertMember(context.Background(), &fleet.FleetMember{ContractID: cid1, FleetID: fleetID, WASMHash: wasm, Active: true})
	_ = repo.UpsertMember(context.Background(), &fleet.FleetMember{ContractID: cid2, FleetID: fleetID, WASMHash: wasm, Active: true})

	mockRes := &mockResolver{
		resolved: map[string]cap85.ResolvedExecutable{
			cid1: {Kind: "EXTERNAL_REF", Fleet: &fleetID, WASMHash: wasm},
			cid2: {Kind: "EXTERNAL_REF", Fleet: &fleetID, WASMHash: wasm},
		},
	}
	mockCp := &mockCheckpoints{cp: 100}
	mockRpcClient := &mockRPC{latestLedger: 100}

	engine := NewEngine(db, repo, mockCp, mockRes, mockRpcClient)
	result, err := engine.VerifyFleet(context.Background(), fleetID, wasm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != StatusHealthy {
		t.Errorf("expected HEALTHY, got %s", result.Status)
	}
	if result.TotalMembers != 2 || result.MatchingMembers != 2 || result.MismatchingMembers != 0 {
		t.Errorf("counts mismatch: %+v", result)
	}
}

func TestVerifier_Drift(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	_, _ = db.Exec("TRUNCATE TABLE fleet_verifications, fleet_members, fleets CASCADE")

	repo := fleet.NewPostgresRepository(db)
	ownerID := makeContractID(2)
	fleetID := FleetID{Owner: ownerID, Tag: "vault-v1"}
	expectedWasm := []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	staleWasm := []byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}

	_ = repo.UpsertFleet(context.Background(), &fleet.Fleet{
		ID:                fleetID,
		CurrentWASMHash:   expectedWasm,
		FirstSeenLedger:   100,
		LastSeenLedger:    100,
		LastIndexedLedger: 100,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	})

	cidA := makeContractID(20)
	cidB := makeContractID(21)
	cidC := makeContractID(22)
	_ = repo.UpsertMember(context.Background(), &fleet.FleetMember{ContractID: cidA, FleetID: fleetID, WASMHash: expectedWasm, Active: true})
	_ = repo.UpsertMember(context.Background(), &fleet.FleetMember{ContractID: cidB, FleetID: fleetID, WASMHash: expectedWasm, Active: true})
	_ = repo.UpsertMember(context.Background(), &fleet.FleetMember{ContractID: cidC, FleetID: fleetID, WASMHash: staleWasm, Active: true})

	mockRes := &mockResolver{
		resolved: map[string]cap85.ResolvedExecutable{
			cidA: {Kind: "EXTERNAL_REF", Fleet: &fleetID, WASMHash: expectedWasm},
			cidB: {Kind: "EXTERNAL_REF", Fleet: &fleetID, WASMHash: expectedWasm},
			cidC: {Kind: "EXTERNAL_REF", Fleet: &fleetID, WASMHash: staleWasm},
		},
	}
	mockCp := &mockCheckpoints{cp: 100}
	mockRpcClient := &mockRPC{latestLedger: 100}

	engine := NewEngine(db, repo, mockCp, mockRes, mockRpcClient)
	result, err := engine.VerifyFleet(context.Background(), fleetID, expectedWasm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Status != StatusDrift {
		t.Errorf("expected DRIFT, got %s", result.Status)
	}
	if result.TotalMembers != 3 || result.MatchingMembers != 2 || result.MismatchingMembers != 1 {
		t.Errorf("counts mismatch for DRIFT: %+v", result)
	}
}

func TestVerifier_BrokenReference(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	_, _ = db.Exec("TRUNCATE TABLE fleet_verifications, fleet_members, fleets CASCADE")

	repo := fleet.NewPostgresRepository(db)
	ownerID := makeContractID(3)
	fleetID := FleetID{Owner: ownerID, Tag: "vault-v1"}
	wasm := []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}

	_ = repo.UpsertFleet(context.Background(), &fleet.Fleet{ID: fleetID, CurrentWASMHash: wasm, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	cid := makeContractID(30)
	_ = repo.UpsertMember(context.Background(), &fleet.FleetMember{ContractID: cid, FleetID: fleetID, WASMHash: wasm, Active: true})

	mockRes := &mockResolver{broken: true}
	mockCp := &mockCheckpoints{cp: 100}
	mockRpcClient := &mockRPC{latestLedger: 100}

	engine := NewEngine(db, repo, mockCp, mockRes, mockRpcClient)
	result, err := engine.VerifyFleet(context.Background(), fleetID, wasm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusBrokenReference {
		t.Errorf("expected BROKEN_REFERENCE, got %s", result.Status)
	}
}

func TestVerifier_Incomplete(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	_, _ = db.Exec("TRUNCATE TABLE fleet_verifications, fleet_members, fleets CASCADE")

	repo := fleet.NewPostgresRepository(db)
	ownerID := makeContractID(4)
	fleetID := FleetID{Owner: ownerID, Tag: "vault-v1"}
	wasm := []byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}

	_ = repo.UpsertFleet(context.Background(), &fleet.Fleet{ID: fleetID, CurrentWASMHash: wasm, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	cid := makeContractID(40)
	_ = repo.UpsertMember(context.Background(), &fleet.FleetMember{ContractID: cid, FleetID: fleetID, WASMHash: wasm, Active: true})

	mockRes := &mockResolver{
		resolved: map[string]cap85.ResolvedExecutable{
			cid: {Kind: "EXTERNAL_REF", Fleet: &fleetID, WASMHash: wasm},
		},
	}
	// Checkpoint is behind network ledger (cp=100, latest=150)
	mockCp := &mockCheckpoints{cp: 100}
	mockRpcClient := &mockRPC{latestLedger: 150}

	engine := NewEngine(db, repo, mockCp, mockRes, mockRpcClient)
	result, err := engine.VerifyFleet(context.Background(), fleetID, wasm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != StatusIncomplete {
		t.Errorf("expected INCOMPLETE due to indexer lag, got %s", result.Status)
	}

	// Also verify: 0 members must NEVER return HEALTHY
	emptyFleetID := FleetID{Owner: makeContractID(5), Tag: "empty"}
	_ = repo.UpsertFleet(context.Background(), &fleet.Fleet{ID: emptyFleetID, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	mockRpcClient.latestLedger = 100 // even if checkpoint is caught up
	emptyResult, err := engine.VerifyFleet(context.Background(), emptyFleetID, wasm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if emptyResult.Status == StatusHealthy {
		t.Errorf("empty fleet must NEVER return HEALTHY, got %s", emptyResult.Status)
	}
}
