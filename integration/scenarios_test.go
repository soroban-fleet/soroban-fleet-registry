package integration

import (
	"context"
	"testing"
	"time"

	"github.com/soroban-fleet/soroban-fleet-registry/internal/cap85"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/fleet"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/history"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/ingest"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/verification"
	"github.com/stellar/go-stellar-sdk/strkey"
	"github.com/stellar/go-stellar-sdk/xdr"
)

func makeID(seed byte) string {
	var b [32]byte
	b[0] = seed
	res, _ := strkey.Encode(strkey.VersionByteContract, b[:])
	return res
}

func makeWasm(seed byte) []byte {
	var b [32]byte
	for i := range b {
		b[i] = seed
	}
	return b[:]
}

func TestSection26_Scenarios(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	_, _ = db.Exec("TRUNCATE TABLE fleet_verifications, fleet_releases, fleet_members, fleets, indexer_checkpoints CASCADE")

	ctx := context.Background()
	fleetRepo := fleet.NewPostgresRepository(db)
	checkpoints := ingest.NewPostgresCheckpointStore(db)

	ownerID := makeID(50)
	tag := "vault-v1"
	fleetID := fleet.FleetID{Owner: ownerID, Tag: tag}
	wasmA := makeWasm(0x11)

	_ = fleetRepo.UpsertFleet(ctx, &fleet.Fleet{
		ID:                fleetID,
		CurrentWASMHash:   wasmA,
		FirstSeenLedger:   100,
		LastSeenLedger:    100,
		LastIndexedLedger: 100,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	})
	_ = checkpoints.SetCheckpointTx(ctx, nil, ingest.DefaultStreamName, 100)

	// Scenario 10: One-member fleet verifies healthy
	singleCID := makeID(51)
	_ = fleetRepo.UpsertMember(ctx, &fleet.FleetMember{ContractID: singleCID, FleetID: fleetID, WASMHash: wasmA, Active: true})

	singleRes := &controllableResolver{
		fallback: cap85.ResolvedExecutable{Kind: "EXTERNAL_REF", Fleet: &fleetID, WASMHash: wasmA},
	}
	singleVerifier := verification.NewEngine(db, fleetRepo, checkpoints, singleRes, &testRPC{currentLedger: 100})
	res1, err := singleVerifier.VerifyFleet(ctx, fleetID, wasmA)
	if err != nil || res1.Status != verification.StatusHealthy || res1.TotalMembers != 1 {
		t.Fatalf("scenario 10 (one-member healthy) failed: status=%s, total=%d, err=%v", res1.Status, res1.TotalMembers, err)
	}

	// Scenario 8: Missing owner data detected as BROKEN_REFERENCE
	brokenRes := &brokenResolver{}
	brokenVerifier := verification.NewEngine(db, fleetRepo, checkpoints, brokenRes, &testRPC{currentLedger: 100})
	resBroken, err := brokenVerifier.VerifyFleet(ctx, fleetID, wasmA)
	if err != nil {
		t.Fatalf("verify error: %v", err)
	}
	if resBroken.Status != verification.StatusBrokenReference {
		t.Errorf("scenario 8: expected BROKEN_REFERENCE, got %s", resBroken.Status)
	}

	// Scenario 16: Incomplete indexing produces INCOMPLETE
	laggingVerifier := verification.NewEngine(db, fleetRepo, checkpoints, singleRes, &testRPC{currentLedger: 200}) // network at 200, indexer at 100
	resIncomplete, err := laggingVerifier.VerifyFleet(ctx, fleetID, wasmA)
	if err != nil || resIncomplete.Status != verification.StatusIncomplete {
		t.Errorf("scenario 16: expected INCOMPLETE due to indexer lag, got %s", resIncomplete.Status)
	}

	// Scenario 17: Malformed input never produces HEALTHY
	badHash := []byte{1, 2, 3} // invalid hash length
	_, err = singleVerifier.VerifyFleet(ctx, fleetID, badHash)
	if err == nil {
		t.Errorf("scenario 17: expected error on malformed expected WASM hash")
	}
}

type brokenResolver struct{}

func (b *brokenResolver) ResolveContractExecutable(ctx context.Context, contractID string) (cap85.ResolvedExecutable, error) {
	return cap85.ResolvedExecutable{}, cap85.ErrBrokenReference
}
func (b *brokenResolver) ResolveContractExecutablesBatch(ctx context.Context, contractIDs []string) (map[string]cap85.ResolvedExecutable, error) {
	return nil, cap85.ErrBrokenReference
}

// TestFactoryPatternDemonstration proves a real-world factory/fleet pattern:
// A factory contract deploys multiple vault instances that all share (factory, "vault-core")
func TestFactoryPatternDemonstration(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	_, _ = db.Exec("TRUNCATE TABLE fleet_verifications, fleet_releases, fleet_members, fleets, indexer_checkpoints CASCADE")

	ctx := context.Background()
	fleetRepo := fleet.NewPostgresRepository(db)
	membership := fleet.NewMembershipManager(fleetRepo)
	checkpoints := ingest.NewPostgresCheckpointStore(db)
	releases := history.NewPostgresRepository(db)

	factoryContract := makeID(80)
	tag := "vault-core"
	coreWasm := makeWasm(0x55)

	factoryFleetID := fleet.FleetID{Owner: factoryContract, Tag: tag}

	// Factory deploys 5 user vault instances
	vaults := make([]string, 5)
	for i := 0; i < 5; i++ {
		vaults[i] = makeID(byte(81 + i))
	}

	factoryScAddr, _ := cap85.AddressToScAddress(factoryContract)
	extRef := xdr.ContractExecutableExternalRef{
		ExecutableOwner: factoryScAddr,
		Tag:             xdr.ScString(tag),
	}
	exec, _ := xdr.NewContractExecutable(xdr.ContractExecutableTypeContractExecutableExternalRef, extRef)

	changes := ingest.LedgerChanges{
		Sequence:      5000,
		CloseTimeUnix: 1700000000,
		TagChanges: []ingest.TagExecutableChange{
			{OwnerAddress: factoryContract, Tag: tag, NewWASMHash: coreWasm, TxHash: "tx_factory_init"},
		},
		ContractChanges: make([]ingest.ContractInstanceChange, 5),
	}
	for i, v := range vaults {
		changes.ContractChanges[i] = ingest.ContractInstanceChange{
			ContractID: v,
			NewExec:    &exec,
			TxHash:     "tx_deploy_vault",
		}
	}

	res := &controllableResolver{
		fallback: cap85.ResolvedExecutable{Kind: "EXTERNAL_REF", Fleet: &factoryFleetID, WASMHash: coreWasm},
	}
	proc := ingest.NewLedgerProcessor(fleetRepo, membership, res, releases)
	tx, _ := db.BeginTx(ctx, nil)
	stats, err := proc.ProcessLedgerTx(ctx, tx, changes)
	if err != nil {
		t.Fatalf("factory deployment error: %v", err)
	}
	_ = checkpoints.SetCheckpointTx(ctx, tx, ingest.DefaultStreamName, 5000)
	_ = tx.Commit()

	if stats.MembersUpdated != 5 {
		t.Fatalf("expected 5 vaults registered, got %d", stats.MembersUpdated)
	}

	// Verify the entire factory fleet
	verifier := verification.NewEngine(db, fleetRepo, checkpoints, res, &testRPC{currentLedger: 5000})
	vResult, err := verifier.VerifyFleet(ctx, factoryFleetID, coreWasm)
	if err != nil {
		t.Fatalf("verify factory fleet: %v", err)
	}
	if vResult.Status != verification.StatusHealthy {
		t.Errorf("expected HEALTHY factory fleet, got %s", vResult.Status)
	}
	if vResult.TotalMembers != 5 || vResult.MatchingMembers != 5 {
		t.Errorf("expected 5/5 matching vaults, got %d/%d", vResult.MatchingMembers, vResult.TotalMembers)
	}
}
