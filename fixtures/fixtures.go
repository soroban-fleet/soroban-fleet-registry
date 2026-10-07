package fixtures

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:embed healthy/fleet_healthy.json
var HealthyFleetJSON []byte

//go:embed drift/fleet_drift.json
var DriftFleetJSON []byte

//go:embed broken-reference/broken_owner_missing.json
var BrokenOwnerMissingJSON []byte

//go:embed broken-reference/broken_invalid_hash.json
var BrokenInvalidHashJSON []byte

//go:embed releases/release_upgrade.json
var ReleaseUpgradeJSON []byte

//go:embed malformed/malformed_tag.json
var MalformedTagJSON []byte

//go:embed malformed/malformed_xdr.bin
var MalformedXDRBin []byte

// FleetFixture represents a test scenario dataset.
type FleetFixture struct {
	OwnerAddress string            `json:"owner_address"`
	Tag          string            `json:"tag"`
	ExpectedWASM string            `json:"expected_wasm"`
	Members      []MemberFixture   `json:"members"`
	Ledger       uint32            `json:"ledger"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

type MemberFixture struct {
	ContractID   string `json:"contract_id"`
	ResolvedWASM string `json:"resolved_wasm"`
	Active       bool   `json:"active"`
}

type ReleaseFixture struct {
	OwnerAddress string `json:"owner_address"`
	Tag          string `json:"tag"`
	OldWASM      string `json:"old_wasm"`
	NewWASM      string `json:"new_wasm"`
	Ledger       uint32 `json:"ledger"`
	TxHash       string `json:"tx_hash"`
}

// LoadHealthyFixture parses the healthy fleet scenario.
func LoadHealthyFixture() (*FleetFixture, error) {
	var f FleetFixture
	if err := json.Unmarshal(HealthyFleetJSON, &f); err != nil {
		return nil, fmt.Errorf("unmarshal healthy fixture: %w", err)
	}
	return &f, nil
}

// LoadDriftFixture parses the drift scenario.
func LoadDriftFixture() (*FleetFixture, error) {
	var f FleetFixture
	if err := json.Unmarshal(DriftFleetJSON, &f); err != nil {
		return nil, fmt.Errorf("unmarshal drift fixture: %w", err)
	}
	return &f, nil
}

// LoadReleaseFixture parses the release upgrade scenario.
func LoadReleaseFixture() (*ReleaseFixture, error) {
	var r ReleaseFixture
	if err := json.Unmarshal(ReleaseUpgradeJSON, &r); err != nil {
		return nil, fmt.Errorf("unmarshal release fixture: %w", err)
	}
	return &r, nil
}
