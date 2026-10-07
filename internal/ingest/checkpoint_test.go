package ingest

import (
	"context"
	"testing"

	"github.com/soroban-fleet/soroban-fleet-registry/internal/cap85"
	"github.com/soroban-fleet/soroban-fleet-registry/internal/fleet"
	"github.com/stellar/go-stellar-sdk/xdr"
)

func TestCheckpointStore_GetSet(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	_, _ = db.Exec("TRUNCATE TABLE indexer_checkpoints CASCADE")

	store := NewPostgresCheckpointStore(db)
	ctx := context.Background()

	// 1. Initial checkpoint must be 0
	cp, err := store.GetCheckpoint(ctx, "test_stream")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cp != 0 {
		t.Errorf("expected 0, got %d", cp)
	}

	// 2. Set checkpoint to 1000 in transaction
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("begin tx: %v", err)
	}
	if err := store.SetCheckpointTx(ctx, tx, "test_stream", 1000); err != nil {
		_ = tx.Rollback()
		t.Fatalf("set checkpoint: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit tx: %v", err)
	}

	// 3. Verify updated checkpoint
	cp, err = store.GetCheckpoint(ctx, "test_stream")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cp != 1000 {
		t.Errorf("expected 1000, got %d", cp)
	}

	// 4. Update checkpoint to 1050
	tx2, _ := db.BeginTx(ctx, nil)
	_ = store.SetCheckpointTx(ctx, tx2, "test_stream", 1050)
	_ = tx2.Commit()

	cp, _ = store.GetCheckpoint(ctx, "test_stream")
	if cp != 1050 {
		t.Errorf("expected 1050, got %d", cp)
	}
}

func TestPipeline_IdempotentReplayAndResume(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	_, _ = db.Exec("TRUNCATE TABLE fleet_members, fleets, indexer_checkpoints CASCADE")

	store := NewPostgresCheckpointStore(db)
	repo := fleet.NewPostgresRepository(db)
	mm := fleet.NewMembershipManager(repo)
	proc := NewLedgerProcessor(repo, mm, nil, nil)
	pipe := NewPipeline(db, nil, proc, store, nil)
	ctx := context.Background()

	stream := "main"
	ownerID := makeContractID(30)
	cid := makeContractID(31)
	tag := "pool-v1"
	wasmHash := [32]byte{0xAA, 0xBB, 0xCC}

	ownerScAddr, _ := cap85.AddressToScAddress(ownerID)
	contractScAddr, _ := cap85.AddressToScAddress(cid)

	extRef := xdr.ContractExecutableExternalRef{
		ExecutableOwner: ownerScAddr,
		Tag:             xdr.ScString(tag),
	}
	exec, _ := xdr.NewContractExecutable(xdr.ContractExecutableTypeContractExecutableExternalRef, extRef)
	instKeyVal, _ := xdr.NewScVal(xdr.ScValTypeScvLedgerKeyContractInstance, nil)
	instVal, _ := xdr.NewScVal(xdr.ScValTypeScvContractInstance, xdr.ScContractInstance{Executable: exec})

	contractEntry := xdr.LedgerEntry{
		LastModifiedLedgerSeq: 500,
		Data: xdr.LedgerEntryData{
			Type: xdr.LedgerEntryTypeContractData,
			ContractData: &xdr.ContractDataEntry{
				Contract:   contractScAddr,
				Key:        instKeyVal,
				Durability: xdr.ContractDataDurabilityPersistent,
				Val:        instVal,
			},
		},
	}

	tagKeyVal, _ := cap85.BuildTagScVal(tag)
	tagVal, _ := xdr.NewScVal(xdr.ScValTypeScvBytes, xdr.ScBytes(wasmHash[:]))
	tagEntry := xdr.LedgerEntry{
		LastModifiedLedgerSeq: 500,
		Data: xdr.LedgerEntryData{
			Type: xdr.LedgerEntryTypeContractData,
			ContractData: &xdr.ContractDataEntry{
				Contract:   ownerScAddr,
				Key:        tagKeyVal,
				Durability: xdr.ContractDataDurabilityPersistent,
				Val:        tagVal,
			},
		},
	}

	changes := xdr.LedgerEntryChanges{
		{Type: xdr.LedgerEntryChangeTypeLedgerEntryCreated, Created: &tagEntry},
		{Type: xdr.LedgerEntryChangeTypeLedgerEntryCreated, Created: &contractEntry},
	}
	txMetaV3 := xdr.TransactionMetaV3{TxChangesAfter: changes}
	txMeta := xdr.TransactionMeta{V: 3, V3: &txMetaV3}
	metaV1 := xdr.LedgerCloseMetaV1{
		LedgerHeader: xdr.LedgerHeaderHistoryEntry{
			Header: xdr.LedgerHeader{
				LedgerSeq: 500,
				ScpValue:  xdr.StellarValue{CloseTime: 1700000000},
			},
		},
		TxProcessing: []xdr.TransactionResultMeta{{TxApplyProcessing: txMeta}},
	}
	meta := xdr.LedgerCloseMeta{V: 1, V1: &metaV1}

	// 1. First processing pass of ledger 500
	stats1, err := pipe.ProcessSingleLedger(ctx, meta, stream)
	if err != nil {
		t.Fatalf("first ledger processing failed: %v", err)
	}
	if stats1.Ledger != 500 {
		t.Errorf("expected ledger 500, got %d", stats1.Ledger)
	}

	cp, err := store.GetCheckpoint(ctx, stream)
	if err != nil || cp != 500 {
		t.Fatalf("checkpoint mismatch after pass 1: got %d, err %v", cp, err)
	}

	fID := fleet.FleetID{Owner: ownerID, Tag: tag}
	f, err := repo.GetFleet(ctx, fID)
	if err != nil || f.MemberCount != 1 {
		t.Fatalf("fleet member count expected 1, got %d (err %v)", f.MemberCount, err)
	}

	var fleetRowCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM fleets").Scan(&fleetRowCount)
	if fleetRowCount != 1 {
		t.Fatalf("expected 1 row in fleets, got %d", fleetRowCount)
	}

	var memberRowCount int
	_ = db.QueryRow("SELECT COUNT(*) FROM fleet_members").Scan(&memberRowCount)
	if memberRowCount != 1 {
		t.Fatalf("expected 1 row in fleet_members, got %d", memberRowCount)
	}

	// 2. IDEMPOTENT REPLAY: Process the EXACT SAME ledger 500 again!
	stats2, err := pipe.ProcessSingleLedger(ctx, meta, stream)
	if err != nil {
		t.Fatalf("second ledger processing (replay) failed: %v", err)
	}
	if stats2.Ledger != 500 {
		t.Errorf("expected ledger 500 on replay, got %d", stats2.Ledger)
	}

	// Verify checkpoint remains 500
	cp, err = store.GetCheckpoint(ctx, stream)
	if err != nil || cp != 500 {
		t.Fatalf("checkpoint after replay: got %d, err %v", cp, err)
	}

	// Verify counts: no duplicate rows in fleets or fleet_members
	_ = db.QueryRow("SELECT COUNT(*) FROM fleets").Scan(&fleetRowCount)
	if fleetRowCount != 1 {
		t.Errorf("duplicate fleet rows detected after replay: got %d", fleetRowCount)
	}

	_ = db.QueryRow("SELECT COUNT(*) FROM fleet_members").Scan(&memberRowCount)
	if memberRowCount != 1 {
		t.Errorf("duplicate member rows detected after replay: got %d", memberRowCount)
	}

	f, _ = repo.GetFleet(ctx, fID)
	if f.MemberCount != 1 {
		t.Errorf("corrupted member count after replay: got %d", f.MemberCount)
	}
}
