package ingest

import (
	"bytes"
	"testing"

	"github.com/soroban-fleet/soroban-fleet-registry/internal/cap85"
	"github.com/stellar/go-stellar-sdk/strkey"
	"github.com/stellar/go-stellar-sdk/xdr"
)

func makeContractID(seed byte) string {
	var b [32]byte
	b[0] = seed
	res, _ := strkey.Encode(strkey.VersionByteContract, b[:])
	return res
}

func TestExtractChangesFromLedgerCloseMeta(t *testing.T) {
	contractID := makeContractID(1)
	ownerID := makeContractID(2)
	tag := "vault-v1"

	contractScAddr, err := cap85.AddressToScAddress(contractID)
	if err != nil {
		t.Fatalf("contract addr: %v", err)
	}
	ownerScAddr, err := cap85.AddressToScAddress(ownerID)
	if err != nil {
		t.Fatalf("owner addr: %v", err)
	}

	// 1. Contract Instance Entry
	instKeyVal, _ := xdr.NewScVal(xdr.ScValTypeScvLedgerKeyContractInstance, nil)
	extRef := xdr.ContractExecutableExternalRef{
		ExecutableOwner: ownerScAddr,
		Tag:             xdr.ScString(tag),
	}
	exec, _ := xdr.NewContractExecutable(xdr.ContractExecutableTypeContractExecutableExternalRef, extRef)
	instVal, _ := xdr.NewScVal(xdr.ScValTypeScvContractInstance, xdr.ScContractInstance{Executable: exec})

	contractEntry := xdr.LedgerEntry{
		LastModifiedLedgerSeq: 100,
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

	// 2. Owner Tag Entry
	tagKeyVal, _ := cap85.BuildTagScVal(tag)
	wasmHash := [32]byte{0xDE, 0xAD, 0xBE, 0xEF}
	tagVal, _ := xdr.NewScVal(xdr.ScValTypeScvBytes, xdr.ScBytes(wasmHash[:]))

	tagEntry := xdr.LedgerEntry{
		LastModifiedLedgerSeq: 100,
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

	// Build LedgerCloseMetaV3 (which has operations and tx changes)
	changes := xdr.LedgerEntryChanges{
		{
			Type:    xdr.LedgerEntryChangeTypeLedgerEntryCreated,
			Created: &contractEntry,
		},
		{
			Type:    xdr.LedgerEntryChangeTypeLedgerEntryCreated,
			Created: &tagEntry,
		},
	}

	txMetaV3 := xdr.TransactionMetaV3{
		TxChangesAfter: changes,
	}
	txMeta := xdr.TransactionMeta{
		V:  3,
		V3: &txMetaV3,
	}

	metaV1 := xdr.LedgerCloseMetaV1{
		LedgerHeader: xdr.LedgerHeaderHistoryEntry{
			Header: xdr.LedgerHeader{
				LedgerSeq: 100,
				ScpValue: xdr.StellarValue{
					CloseTime: 1700000000,
				},
			},
		},
		TxProcessing: []xdr.TransactionResultMeta{
			{
				TxApplyProcessing: txMeta,
			},
		},
	}

	meta := xdr.LedgerCloseMeta{
		V:  1,
		V1: &metaV1,
	}

	extracted, err := ExtractChangesFromLedgerCloseMeta(meta)
	if err != nil {
		t.Fatalf("unexpected error extracting changes: %v", err)
	}

	if extracted.Sequence != 100 {
		t.Errorf("expected sequence 100, got %d", extracted.Sequence)
	}
	if len(extracted.ContractChanges) != 1 {
		t.Fatalf("expected 1 contract change, got %d", len(extracted.ContractChanges))
	}
	if extracted.ContractChanges[0].ContractID != contractID {
		t.Errorf("contract ID mismatch: %s vs %s", extracted.ContractChanges[0].ContractID, contractID)
	}

	if len(extracted.TagChanges) != 1 {
		t.Fatalf("expected 1 tag change, got %d", len(extracted.TagChanges))
	}
	tc := extracted.TagChanges[0]
	if tc.OwnerAddress != ownerID || tc.Tag != tag {
		t.Errorf("tag change mismatch: %+v", tc)
	}
	if !bytes.Equal(tc.NewWASMHash, wasmHash[:]) {
		t.Errorf("wasm hash mismatch: %x vs %x", tc.NewWASMHash, wasmHash[:])
	}
}
