package cap85

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/soroban-fleet/soroban-fleet-registry/internal/rpc"
	protocol "github.com/stellar/go-stellar-sdk/protocols/rpc"
	"github.com/stellar/go-stellar-sdk/strkey"
	"github.com/stellar/go-stellar-sdk/xdr"
)

type mockRPC struct {
	entries map[string]protocol.LedgerEntryResult
}

func newMockRPC() *mockRPC {
	return &mockRPC{
		entries: make(map[string]protocol.LedgerEntryResult),
	}
}

func (m *mockRPC) addContractInstance(contractID string, exec xdr.ContractExecutable) error {
	scAddr, err := AddressToScAddress(contractID)
	if err != nil {
		return err
	}
	keyVal, err := xdr.NewScVal(xdr.ScValTypeScvLedgerKeyContractInstance, nil)
	if err != nil {
		return err
	}
	var key xdr.LedgerKey
	if err := key.SetContractData(scAddr, keyVal, xdr.ContractDataDurabilityPersistent); err != nil {
		return err
	}
	keyBytes, err := key.MarshalBinary()
	if err != nil {
		return err
	}
	keyB64 := base64.StdEncoding.EncodeToString(keyBytes)

	instVal, err := xdr.NewScVal(xdr.ScValTypeScvContractInstance, xdr.ScContractInstance{
		Executable: exec,
	})
	if err != nil {
		return err
	}

	entryData := xdr.LedgerEntryData{
		Type: xdr.LedgerEntryTypeContractData,
		ContractData: &xdr.ContractDataEntry{
			Contract:   scAddr,
			Key:        keyVal,
			Durability: xdr.ContractDataDurabilityPersistent,
			Val:        instVal,
		},
	}
	dataBytes, err := entryData.MarshalBinary()
	if err != nil {
		return err
	}
	m.entries[keyB64] = protocol.LedgerEntryResult{
		KeyXDR:             keyB64,
		DataXDR:            base64.StdEncoding.EncodeToString(dataBytes),
		LastModifiedLedger: 100,
	}
	return nil
}

func (m *mockRPC) addOwnerTagEntry(ownerID string, tag string, wasmHash []byte) error {
	ownerAddr, err := AddressToScAddress(ownerID)
	if err != nil {
		return err
	}
	tagVal, err := BuildTagScVal(tag)
	if err != nil {
		return err
	}
	var key xdr.LedgerKey
	if err := key.SetContractData(ownerAddr, tagVal, xdr.ContractDataDurabilityPersistent); err != nil {
		return err
	}
	keyBytes, err := key.MarshalBinary()
	if err != nil {
		return err
	}
	keyB64 := base64.StdEncoding.EncodeToString(keyBytes)

	hashVal, err := xdr.NewScVal(xdr.ScValTypeScvBytes, xdr.ScBytes(wasmHash))
	if err != nil {
		return err
	}

	entryData := xdr.LedgerEntryData{
		Type: xdr.LedgerEntryTypeContractData,
		ContractData: &xdr.ContractDataEntry{
			Contract:   ownerAddr,
			Key:        tagVal,
			Durability: xdr.ContractDataDurabilityPersistent,
			Val:        hashVal,
		},
	}
	dataBytes, err := entryData.MarshalBinary()
	if err != nil {
		return err
	}
	m.entries[keyB64] = protocol.LedgerEntryResult{
		KeyXDR:             keyB64,
		DataXDR:            base64.StdEncoding.EncodeToString(dataBytes),
		LastModifiedLedger: 100,
	}
	return nil
}

func (m *mockRPC) GetLedgerEntries(ctx context.Context, keys []xdr.LedgerKey) (rpc.LedgerEntriesResponse, error) {
	var results []protocol.LedgerEntryResult
	for _, k := range keys {
		raw, err := k.MarshalBinary()
		if err != nil {
			return rpc.LedgerEntriesResponse{}, err
		}
		b64 := base64.StdEncoding.EncodeToString(raw)
		if res, ok := m.entries[b64]; ok {
			results = append(results, res)
		}
	}
	return rpc.LedgerEntriesResponse{
		Entries:      results,
		LatestLedger: 200,
	}, nil
}

func (m *mockRPC) GetLatestLedger(ctx context.Context) (rpc.LatestLedger, error) {
	return rpc.LatestLedger{
		Sequence: 200,
	}, nil
}

func (m *mockRPC) GetLedgers(ctx context.Context, startLedger uint32, limit uint32) (protocol.GetLedgersResponse, error) {
	return protocol.GetLedgersResponse{}, nil
}

func makeContractID(seed byte) string {
	var b [32]byte
	b[0] = seed
	res, _ := strkey.Encode(strkey.VersionByteContract, b[:])
	return res
}

func TestResolveContractExecutable_DirectWasm(t *testing.T) {
	mock := newMockRPC()
	cid := makeContractID(1)

	var wasmHash [32]byte
	wasmHash[0] = 0xAA
	exec, err := xdr.NewContractExecutable(xdr.ContractExecutableTypeContractExecutableWasm, xdr.Hash(wasmHash))
	if err != nil {
		t.Fatalf("failed to create wasm exec: %v", err)
	}

	if err := mock.addContractInstance(cid, exec); err != nil {
		t.Fatalf("failed to add contract instance: %v", err)
	}

	resolver := NewResolver(mock)
	res, err := resolver.ResolveContractExecutable(context.Background(), cid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Kind != "WASM" {
		t.Errorf("expected Kind WASM, got %s", res.Kind)
	}
	if !bytes.Equal(res.WASMHash, wasmHash[:]) {
		t.Errorf("expected wasm hash %x, got %x", wasmHash, res.WASMHash)
	}
	if res.Fleet != nil {
		t.Errorf("expected nil Fleet for direct wasm")
	}
}

func TestResolveContractExecutable_ExternalRef(t *testing.T) {
	mock := newMockRPC()
	cid := makeContractID(2)
	ownerID := makeContractID(3)
	tag := "vault-v1"

	var expectedHash [32]byte
	expectedHash[0] = 0xBB
	expectedHash[31] = 0xCC

	ownerAddr, err := AddressToScAddress(ownerID)
	if err != nil {
		t.Fatalf("failed to parse owner address: %v", err)
	}

	extRef := xdr.ContractExecutableExternalRef{
		ExecutableOwner: ownerAddr,
		Tag:             xdr.ScString(tag),
	}
	exec, err := xdr.NewContractExecutable(xdr.ContractExecutableTypeContractExecutableExternalRef, extRef)
	if err != nil {
		t.Fatalf("failed to create external ref exec: %v", err)
	}

	if err := mock.addContractInstance(cid, exec); err != nil {
		t.Fatalf("failed to add contract instance: %v", err)
	}
	if err := mock.addOwnerTagEntry(ownerID, tag, expectedHash[:]); err != nil {
		t.Fatalf("failed to add owner tag entry: %v", err)
	}

	resolver := NewResolver(mock)
	res, err := resolver.ResolveContractExecutable(context.Background(), cid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if res.Kind != "EXTERNAL_REF" {
		t.Errorf("expected Kind EXTERNAL_REF, got %s", res.Kind)
	}
	if res.Fleet == nil {
		t.Fatalf("expected non-nil Fleet")
	}
	if res.Fleet.Owner != ownerID || res.Fleet.Tag != tag {
		t.Errorf("expected fleet %+v, got %+v", FleetID{Owner: ownerID, Tag: tag}, *res.Fleet)
	}
	if !bytes.Equal(res.WASMHash, expectedHash[:]) {
		t.Errorf("expected wasm hash %x, got %x", expectedHash, res.WASMHash)
	}
}

func TestResolveContractExecutable_BrokenReference(t *testing.T) {
	mock := newMockRPC()
	cid := makeContractID(4)
	ownerID := makeContractID(5)
	tag := "vault-broken"

	ownerAddr, _ := AddressToScAddress(ownerID)
	extRef := xdr.ContractExecutableExternalRef{
		ExecutableOwner: ownerAddr,
		Tag:             xdr.ScString(tag),
	}
	exec, _ := xdr.NewContractExecutable(xdr.ContractExecutableTypeContractExecutableExternalRef, extRef)
	_ = mock.addContractInstance(cid, exec)

	// Note: owner tag entry NOT added, so reference is broken
	resolver := NewResolver(mock)
	_, err := resolver.ResolveContractExecutable(context.Background(), cid)
	if !errors.Is(err, ErrBrokenReference) {
		t.Fatalf("expected ErrBrokenReference when owner tag entry is missing, got %v", err)
	}

	// Missing contract instance
	missingCID := makeContractID(99)
	_, err = resolver.ResolveContractExecutable(context.Background(), missingCID)
	if !errors.Is(err, ErrBrokenReference) {
		t.Fatalf("expected ErrBrokenReference when contract instance is missing, got %v", err)
	}
}

func TestResolveContractExecutablesBatch(t *testing.T) {
	mock := newMockRPC()
	cidWasm := makeContractID(10)
	cidExt := makeContractID(11)
	ownerID := makeContractID(12)
	tag := "token-pool"

	var hashWasm [32]byte
	hashWasm[0] = 0x11
	execWasm, _ := xdr.NewContractExecutable(xdr.ContractExecutableTypeContractExecutableWasm, xdr.Hash(hashWasm))
	_ = mock.addContractInstance(cidWasm, execWasm)

	var hashExt [32]byte
	hashExt[0] = 0x22
	ownerAddr, _ := AddressToScAddress(ownerID)
	execExt, _ := xdr.NewContractExecutable(xdr.ContractExecutableTypeContractExecutableExternalRef, xdr.ContractExecutableExternalRef{
		ExecutableOwner: ownerAddr,
		Tag:             xdr.ScString(tag),
	})
	_ = mock.addContractInstance(cidExt, execExt)
	_ = mock.addOwnerTagEntry(ownerID, tag, hashExt[:])

	resolver := NewResolver(mock)
	results, err := resolver.ResolveContractExecutablesBatch(context.Background(), []string{cidWasm, cidExt})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	if results[cidWasm].Kind != "WASM" || !bytes.Equal(results[cidWasm].WASMHash, hashWasm[:]) {
		t.Errorf("wasm contract result mismatch: %+v", results[cidWasm])
	}

	if results[cidExt].Kind != "EXTERNAL_REF" || !bytes.Equal(results[cidExt].WASMHash, hashExt[:]) {
		t.Errorf("ext ref contract result mismatch: %+v", results[cidExt])
	}
}
