package fleet

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/soroban-fleet/soroban-fleet-registry/internal/cap85"
)

// MembershipManager coordinates member lifecycle changes within database transactions.
type MembershipManager struct {
	repo Repository
}

// NewMembershipManager creates a new MembershipManager.
func NewMembershipManager(repo Repository) *MembershipManager {
	return &MembershipManager{repo: repo}
}

// RegisterOrUpdateMemberTx registers or updates a contract as an active fleet member.
// If the contract previously belonged to a different fleet, its old membership is set to inactive.
func (m *MembershipManager) RegisterOrUpdateMemberTx(
	ctx context.Context,
	tx *sql.Tx,
	contractID string,
	fleetID FleetID,
	wasmHash []byte,
	ledger uint32,
) error {
	existing, err := m.repo.GetMember(ctx, contractID)
	if err != nil && !errorsIs(err, ErrFleetMemberNotFound) {
		return fmt.Errorf("check existing member %s: %w", contractID, err)
	}

	// If contract existed and was in a different fleet, deactivate old fleet membership
	if existing != nil && (existing.FleetID.Owner != fleetID.Owner || existing.FleetID.Tag != fleetID.Tag) {
		if err := m.repo.SetMemberActiveTx(ctx, tx, contractID, false, ledger); err != nil {
			return fmt.Errorf("deactivate old fleet membership: %w", err)
		}
		if err := m.repo.RefreshFleetMemberCountTx(ctx, tx, existing.FleetID); err != nil {
			return fmt.Errorf("refresh old fleet member count: %w", err)
		}
	}

	firstSeen := ledger
	if existing != nil {
		firstSeen = existing.FirstSeenLedger
	}

	member := &FleetMember{
		ContractID:      contractID,
		FleetID:         fleetID,
		WASMHash:        wasmHash,
		FirstSeenLedger: firstSeen,
		LastSeenLedger:  ledger,
		Active:          true,
	}

	if err := m.repo.UpsertMemberTx(ctx, tx, member); err != nil {
		return fmt.Errorf("upsert member %s: %w", contractID, err)
	}

	if err := m.repo.RefreshFleetMemberCountTx(ctx, tx, fleetID); err != nil {
		return fmt.Errorf("refresh fleet member count: %w", err)
	}

	return nil
}

// DeactivateMemberTx marks a contract as inactive in its current fleet.
func (m *MembershipManager) DeactivateMemberTx(
	ctx context.Context,
	tx *sql.Tx,
	contractID string,
	ledger uint32,
) error {
	existing, err := m.repo.GetMember(ctx, contractID)
	if err != nil {
		if errorsIs(err, ErrFleetMemberNotFound) {
			return nil // Nothing to deactivate
		}
		return fmt.Errorf("get member %s for deactivation: %w", contractID, err)
	}

	if !existing.Active {
		return nil // Already inactive
	}

	if err := m.repo.SetMemberActiveTx(ctx, tx, contractID, false, ledger); err != nil {
		return fmt.Errorf("set member active false: %w", err)
	}

	if err := m.repo.RefreshFleetMemberCountTx(ctx, tx, existing.FleetID); err != nil {
		return fmt.Errorf("refresh member count after deactivation: %w", err)
	}

	return nil
}

func errorsIs(err, target error) bool {
	if err == nil {
		return target == nil
	}
	return err == target || (cap85.ErrInvalidContractID != nil && err.Error() == target.Error())
}
