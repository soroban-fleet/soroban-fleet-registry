package fleet

import (
	"errors"
	"time"

	"github.com/soroban-fleet/soroban-fleet-registry/internal/cap85"
)

var (
	// ErrFleetNotFound indicates the fleet does not exist.
	ErrFleetNotFound = errors.New("fleet not found")

	// ErrFleetMemberNotFound indicates the contract is not a fleet member.
	ErrFleetMemberNotFound = errors.New("fleet member not found")
)

// FleetID uniquely identifies a fleet by owner and tag.
type FleetID = cap85.FleetID

// Fleet represents an externally managed Soroban contract fleet.
type Fleet struct {
	ID                FleetID   `json:"id"`
	CurrentWASMHash   []byte    `json:"current_wasm_hash"`
	MemberCount       int64     `json:"member_count"`
	FirstSeenLedger   uint32    `json:"first_seen_ledger"`
	LastSeenLedger    uint32    `json:"last_seen_ledger"`
	LastIndexedLedger uint32    `json:"last_indexed_ledger"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// FleetMember represents an individual contract instance belonging to a fleet.
type FleetMember struct {
	ContractID         string  `json:"contract_id"`
	FleetID            FleetID `json:"fleet_id"`
	WASMHash           []byte  `json:"wasm_hash"`
	FirstSeenLedger    uint32  `json:"first_seen_ledger"`
	LastSeenLedger     uint32  `json:"last_seen_ledger"`
	LastVerifiedLedger *uint32 `json:"last_verified_ledger,omitempty"`
	Active             bool    `json:"active"`
}
