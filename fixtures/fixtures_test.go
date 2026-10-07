package fixtures

import (
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stellar/go-stellar-sdk/strkey"
)

func makeContract(seed byte) string {
	var b [32]byte
	b[0] = seed
	res, _ := strkey.Encode(strkey.VersionByteContract, b[:])
	return res
}

func makeHash(seed byte) string {
	var b [32]byte
	for i := range b {
		b[i] = seed
	}
	return hex.EncodeToString(b[:])
}

func TestGenerateAndVerifyFixtures(t *testing.T) {
	ownerID := makeContract(1)
	contractA := makeContract(2)
	contractB := makeContract(3)
	contractC := makeContract(4)

	wasmHashA := makeHash(0xAA)
	wasmHashB := makeHash(0xBB)

	// 1. Healthy fixture
	healthy := FleetFixture{
		OwnerAddress: ownerID,
		Tag:          "vault-v1",
		ExpectedWASM: wasmHashA,
		Ledger:       1000,
		Members: []MemberFixture{
			{ContractID: contractA, ResolvedWASM: wasmHashA, Active: true},
			{ContractID: contractB, ResolvedWASM: wasmHashA, Active: true},
			{ContractID: contractC, ResolvedWASM: wasmHashA, Active: true},
		},
		Metadata: map[string]string{
			"scenario": "flagship_initial_healthy",
		},
	}
	hBytes, _ := json.MarshalIndent(healthy, "", "  ")
	_ = os.WriteFile(filepath.Join("healthy", "fleet_healthy.json"), hBytes, 0644)

	// 2. Drift fixture (Contract C has stale wasmHashA, while expected is wasmHashB)
	drift := FleetFixture{
		OwnerAddress: ownerID,
		Tag:          "vault-v1",
		ExpectedWASM: wasmHashB,
		Ledger:       1100,
		Members: []MemberFixture{
			{ContractID: contractA, ResolvedWASM: wasmHashB, Active: true},
			{ContractID: contractB, ResolvedWASM: wasmHashB, Active: true},
			{ContractID: contractC, ResolvedWASM: wasmHashA, Active: true}, // Stale!
		},
		Metadata: map[string]string{
			"scenario": "flagship_drift_one_stale_member",
		},
	}
	dBytes, _ := json.MarshalIndent(drift, "", "  ")
	_ = os.WriteFile(filepath.Join("drift", "fleet_drift.json"), dBytes, 0644)

	// 3. Broken reference (missing owner tag entry)
	brokenMissing := FleetFixture{
		OwnerAddress: makeContract(5),
		Tag:          "vault-broken",
		ExpectedWASM: wasmHashA,
		Ledger:       1000,
		Members: []MemberFixture{
			{ContractID: makeContract(6), ResolvedWASM: "", Active: true},
		},
		Metadata: map[string]string{
			"scenario": "broken_reference_missing_owner_entry",
		},
	}
	bmBytes, _ := json.MarshalIndent(brokenMissing, "", "  ")
	_ = os.WriteFile(filepath.Join("broken-reference", "broken_owner_missing.json"), bmBytes, 0644)

	// 4. Broken reference (invalid hash length)
	brokenInvalid := FleetFixture{
		OwnerAddress: makeContract(7),
		Tag:          "vault-invalid-hash",
		ExpectedWASM: "1234", // invalid length
		Ledger:       1000,
		Members: []MemberFixture{
			{ContractID: makeContract(8), ResolvedWASM: "1234", Active: true},
		},
		Metadata: map[string]string{
			"scenario": "broken_reference_invalid_hash_length",
		},
	}
	biBytes, _ := json.MarshalIndent(brokenInvalid, "", "  ")
	_ = os.WriteFile(filepath.Join("broken-reference", "broken_invalid_hash.json"), biBytes, 0644)

	// 5. Release fixture
	release := ReleaseFixture{
		OwnerAddress: ownerID,
		Tag:          "vault-v1",
		OldWASM:      wasmHashA,
		NewWASM:      wasmHashB,
		Ledger:       1050,
		TxHash:       "4a72d3f8e0b1c2a3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b",
	}
	relBytes, _ := json.MarshalIndent(release, "", "  ")
	_ = os.WriteFile(filepath.Join("releases", "release_upgrade.json"), relBytes, 0644)

	// 6. Malformed tag fixture
	malformedTag := map[string]any{
		"invalid_tags": []string{
			"",
			string([]byte{0xff, 0xfe, 0xfd}),
		},
	}
	mtBytes, _ := json.MarshalIndent(malformedTag, "", "  ")
	_ = os.WriteFile(filepath.Join("malformed", "malformed_tag.json"), mtBytes, 0644)

	// 7. Malformed XDR binary
	_ = os.WriteFile(filepath.Join("malformed", "malformed_xdr.bin"), []byte{0x00, 0x01, 0x02, 0x03}, 0644)

	// Verify loaders work
	loadedH, err := LoadHealthyFixture()
	if err != nil {
		t.Fatalf("failed to load healthy fixture: %v", err)
	}
	if len(loadedH.Members) != 3 {
		t.Errorf("expected 3 members in healthy fixture, got %d", len(loadedH.Members))
	}

	loadedD, err := LoadDriftFixture()
	if err != nil {
		t.Fatalf("failed to load drift fixture: %v", err)
	}
	if len(loadedD.Members) != 3 {
		t.Errorf("expected 3 members in drift fixture, got %d", len(loadedD.Members))
	}

	loadedR, err := LoadReleaseFixture()
	if err != nil {
		t.Fatalf("failed to load release fixture: %v", err)
	}
	if loadedR.OldWASM != wasmHashA || loadedR.NewWASM != wasmHashB {
		t.Errorf("release fixture hashes mismatch")
	}
}
