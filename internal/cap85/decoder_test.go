package cap85

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stellar/go-stellar-sdk/strkey"
	"github.com/stellar/go-stellar-sdk/xdr"
)

func TestDecodeExternalExecutableRef_Valid(t *testing.T) {
	contractBytes := [32]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32}
	expectedContractStr, err := strkey.Encode(strkey.VersionByteContract, contractBytes[:])
	if err != nil {
		t.Fatalf("failed to encode contract address: %v", err)
	}

	ownerScAddr, err := xdr.NewScAddress(xdr.ScAddressTypeScAddressTypeContract, xdr.ContractId(contractBytes))
	if err != nil {
		t.Fatalf("failed to create ScAddress: %v", err)
	}

	tag := "vault-v1"
	extRef := xdr.ContractExecutableExternalRef{
		ExecutableOwner: ownerScAddr,
		Tag:             xdr.ScString(tag),
	}

	executable, err := xdr.NewContractExecutable(xdr.ContractExecutableTypeContractExecutableExternalRef, extRef)
	if err != nil {
		t.Fatalf("failed to create ContractExecutable: %v", err)
	}

	// Test decoding
	decoded, isExternal, err := DecodeExternalExecutableRef(executable)
	if err != nil {
		t.Fatalf("unexpected error decoding valid external ref: %v", err)
	}
	if !isExternal {
		t.Fatalf("expected isExternal=true")
	}
	if decoded.Owner != expectedContractStr {
		t.Errorf("expected owner %s, got %s", expectedContractStr, decoded.Owner)
	}
	if decoded.Tag != tag {
		t.Errorf("expected tag %s, got %s", tag, decoded.Tag)
	}

	// Test fixture serialization and roundtrip
	fixtureBytes, err := executable.MarshalBinary()
	if err != nil {
		t.Fatalf("failed to marshal executable to binary: %v", err)
	}

	fixturePath := filepath.Join("..", "..", "fixtures", "healthy", "external_ref.xdr")
	if err := os.MkdirAll(filepath.Dir(fixturePath), 0755); err != nil {
		t.Fatalf("failed to create fixtures directory: %v", err)
	}
	if err := os.WriteFile(fixturePath, fixtureBytes, 0644); err != nil {
		t.Fatalf("failed to write fixture file: %v", err)
	}

	// Read fixture file back and unmarshal
	readBytes, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("failed to read fixture file: %v", err)
	}
	var unmarshaled xdr.ContractExecutable
	if err := unmarshaled.UnmarshalBinary(readBytes); err != nil {
		t.Fatalf("failed to unmarshal fixture binary: %v", err)
	}

	decodedFromFixture, isExtFromFixture, err := DecodeExternalExecutableRef(unmarshaled)
	if err != nil {
		t.Fatalf("failed to decode unmarshaled fixture: %v", err)
	}
	if !isExtFromFixture {
		t.Fatalf("expected isExtFromFixture=true")
	}
	if decodedFromFixture.Owner != expectedContractStr {
		t.Errorf("expected fixture owner %s, got %s", expectedContractStr, decodedFromFixture.Owner)
	}
	if decodedFromFixture.Tag != tag {
		t.Errorf("expected fixture tag %s, got %s", tag, decodedFromFixture.Tag)
	}
}

func TestDecodeExternalExecutableRef_Wasm(t *testing.T) {
	wasmHash := xdr.Hash([32]byte{42, 42, 42})
	executable, err := xdr.NewContractExecutable(xdr.ContractExecutableTypeContractExecutableWasm, wasmHash)
	if err != nil {
		t.Fatalf("failed to create wasm executable: %v", err)
	}

	decoded, isExternal, err := DecodeExternalExecutableRef(executable)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if isExternal {
		t.Fatalf("expected isExternal=false for wasm executable")
	}
	if decoded != (ExternalExecutableRef{}) {
		t.Fatalf("expected zero ExternalExecutableRef, got %+v", decoded)
	}
}

func TestDecodeExternalExecutableRef_StellarAsset(t *testing.T) {
	executable, err := xdr.NewContractExecutable(xdr.ContractExecutableTypeContractExecutableStellarAsset, nil)
	if err != nil {
		t.Fatalf("failed to create stellar asset executable: %v", err)
	}

	decoded, isExternal, err := DecodeExternalExecutableRef(executable)
	if err != nil {
		t.Fatalf("expected nil error, got: %v", err)
	}
	if isExternal {
		t.Fatalf("expected isExternal=false for stellar asset executable")
	}
	if decoded != (ExternalExecutableRef{}) {
		t.Fatalf("expected zero ExternalExecutableRef, got %+v", decoded)
	}
}

func TestDecodeExternalExecutableRef_Malformed(t *testing.T) {
	// Missing external ref in union
	executable := xdr.ContractExecutable{
		Type:        xdr.ContractExecutableTypeContractExecutableExternalRef,
		ExternalRef: nil,
	}
	_, _, err := DecodeExternalExecutableRef(executable)
	if !errors.Is(err, ErrInvalidExternalRef) {
		t.Fatalf("expected ErrInvalidExternalRef for nil external_ref, got %v", err)
	}

	// Empty tag in external ref
	addr, err := xdr.NewScAddress(xdr.ScAddressTypeScAddressTypeContract, xdr.ContractId([32]byte{1}))
	if err != nil {
		t.Fatalf("failed to create ScAddress: %v", err)
	}
	badTagExecutable := xdr.ContractExecutable{
		Type: xdr.ContractExecutableTypeContractExecutableExternalRef,
		ExternalRef: &xdr.ContractExecutableExternalRef{
			ExecutableOwner: addr,
			Tag:             "",
		},
	}
	_, _, err = DecodeExternalExecutableRef(badTagExecutable)
	if !errors.Is(err, ErrInvalidExternalRef) {
		t.Fatalf("expected ErrInvalidExternalRef for empty tag, got %v", err)
	}

	// Unsupported executable type
	badTypeExecutable := xdr.ContractExecutable{
		Type: xdr.ContractExecutableType(999),
	}
	_, _, err = DecodeExternalExecutableRef(badTypeExecutable)
	if !errors.Is(err, ErrInvalidExternalRef) {
		t.Fatalf("expected ErrInvalidExternalRef for unknown type, got %v", err)
	}
}

func TestAddressToScAddressAndDecodeScAddress(t *testing.T) {
	// Test Contract address roundtrip
	contractBytes := [32]byte{10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 110, 120, 130, 140, 150, 160, 170, 180, 190, 200, 210, 220, 230, 240, 250, 1, 2, 3, 4, 5, 6, 7}
	contractStr, err := strkey.Encode(strkey.VersionByteContract, contractBytes[:])
	if err != nil {
		t.Fatalf("failed to encode contract strkey: %v", err)
	}

	scAddr, err := AddressToScAddress(contractStr)
	if err != nil {
		t.Fatalf("failed to parse contract address: %v", err)
	}
	if scAddr.Type != xdr.ScAddressTypeScAddressTypeContract {
		t.Fatalf("expected contract ScAddress type, got %v", scAddr.Type)
	}
	cid := scAddr.MustContractId()
	if !bytes.Equal(cid[:], contractBytes[:]) {
		t.Fatalf("contract ID bytes mismatch")
	}

	roundtripContract, err := DecodeScAddress(scAddr)
	if err != nil {
		t.Fatalf("failed to decode ScAddress: %v", err)
	}
	if roundtripContract != contractStr {
		t.Fatalf("expected %s, got %s", contractStr, roundtripContract)
	}

	// Test Account address roundtrip
	accountBytes := [32]byte{7, 6, 5, 4, 3, 2, 1, 250, 240, 230, 220, 210, 200, 190, 180, 170, 160, 150, 140, 130, 120, 110, 100, 90, 80, 70, 60, 50, 40, 30, 20, 10}
	accountStr, err := strkey.Encode(strkey.VersionByteAccountID, accountBytes[:])
	if err != nil {
		t.Fatalf("failed to encode account strkey: %v", err)
	}

	scAccountAddr, err := AddressToScAddress(accountStr)
	if err != nil {
		t.Fatalf("failed to parse account address: %v", err)
	}
	if scAccountAddr.Type != xdr.ScAddressTypeScAddressTypeAccount {
		t.Fatalf("expected account ScAddress type, got %v", scAccountAddr.Type)
	}
	if !bytes.Equal(scAccountAddr.MustAccountId().Ed25519[:], accountBytes[:]) {
		t.Fatalf("account ID bytes mismatch")
	}

	roundtripAccount, err := DecodeScAddress(scAccountAddr)
	if err != nil {
		t.Fatalf("failed to decode ScAddress: %v", err)
	}
	if roundtripAccount != accountStr {
		t.Fatalf("expected %s, got %s", accountStr, roundtripAccount)
	}

	// Test error cases
	if _, err := AddressToScAddress(""); !errors.Is(err, ErrInvalidContractID) {
		t.Errorf("expected ErrInvalidContractID for empty address, got %v", err)
	}
	if _, err := AddressToScAddress("invalid_strkey"); !errors.Is(err, ErrInvalidContractID) {
		t.Errorf("expected ErrInvalidContractID for invalid strkey, got %v", err)
	}
}
