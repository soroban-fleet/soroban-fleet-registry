package verification

import (
	"time"

	"github.com/soroban-fleet/soroban-fleet-registry/internal/cap85"
)

// FleetID uniquely identifies a fleet.
type FleetID = cap85.FleetID

// VerificationStatus represents the classification of a fleet's current integrity.
type VerificationStatus string

const (
	// StatusHealthy: all active members resolve, all match expected hash, coverage is complete.
	StatusHealthy VerificationStatus = "HEALTHY"

	// StatusDrift: one or more active members resolve to a different WASM.
	StatusDrift VerificationStatus = "DRIFT"

	// StatusBrokenReference: the fleet's external reference cannot resolve.
	StatusBrokenReference VerificationStatus = "BROKEN_REFERENCE"

	// StatusIncomplete: indexer has not reached the height required for complete assertion.
	StatusIncomplete VerificationStatus = "INCOMPLETE"

	// StatusUnknown: insufficient information to safely classify.
	StatusUnknown VerificationStatus = "UNKNOWN"
)

// VerificationResult contains the detailed findings of a fleet verification.
type VerificationResult struct {
	ID                 int64              `json:"id,omitempty"`
	FleetID            FleetID            `json:"fleet_id"`
	ExpectedWASMHash   []byte             `json:"expected_wasm_hash"`
	TotalMembers       int64              `json:"total_members"`
	MatchingMembers    int64              `json:"matching_members"`
	MismatchingMembers int64              `json:"mismatching_members"`
	MissingMembers     int64              `json:"missing_members"`
	IndexedThrough     uint32             `json:"indexed_through"`
	Status             VerificationStatus `json:"status"`
	VerifiedAt         time.Time          `json:"verified_at"`
}
