package fleet

import (
	"context"
	"fmt"

	"github.com/soroban-fleet/soroban-fleet-registry/internal/cap85"
	"github.com/stellar/go-stellar-sdk/xdr"
)

// ContractInspection represents the complete indexed and resolved state of a contract.
type ContractInspection struct {
	ContractID string                   `json:"contract_id"`
	IsMember   bool                     `json:"is_member"`
	Member     *FleetMember             `json:"member,omitempty"`
	Resolved   cap85.ResolvedExecutable `json:"resolved"`
}

// Service provides high-level fleet inspection and query operations.
type Service struct {
	repo     Repository
	resolver cap85.Resolver
}

// NewService creates a new fleet Service.
func NewService(repo Repository, resolver cap85.Resolver) *Service {
	return &Service{
		repo:     repo,
		resolver: resolver,
	}
}

// GetFleet fetches a fleet after validating its identifier.
func (s *Service) GetFleet(ctx context.Context, id FleetID) (*Fleet, error) {
	if err := s.validateFleetID(id); err != nil {
		return nil, err
	}
	return s.repo.GetFleet(ctx, id)
}

// ListFleets returns a paginated list of fleets.
func (s *Service) ListFleets(ctx context.Context, limit, offset int) ([]Fleet, int64, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListFleets(ctx, limit, offset)
}

// ListMembers returns paginated members of a fleet.
func (s *Service) ListMembers(ctx context.Context, fleetID FleetID, activeOnly bool, limit, offset int) ([]FleetMember, int64, error) {
	if err := s.validateFleetID(fleetID); err != nil {
		return nil, 0, err
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return s.repo.ListMembers(ctx, fleetID, activeOnly, limit, offset)
}

// GetMember retrieves member details by contract ID.
func (s *Service) GetMember(ctx context.Context, contractID string) (*FleetMember, error) {
	if _, err := cap85.AddressToScAddress(contractID); err != nil {
		return nil, fmt.Errorf("%w: %s", cap85.ErrInvalidContractID, contractID)
	}
	return s.repo.GetMember(ctx, contractID)
}

// InspectContract returns both indexed fleet membership (if any) and live resolved executable state.
func (s *Service) InspectContract(ctx context.Context, contractID string) (*ContractInspection, error) {
	if _, err := cap85.AddressToScAddress(contractID); err != nil {
		return nil, fmt.Errorf("%w: %s", cap85.ErrInvalidContractID, contractID)
	}

	var member *FleetMember
	isMember := false

	m, err := s.repo.GetMember(ctx, contractID)
	if err == nil {
		member = m
		isMember = m.Active
	}

	var resolved cap85.ResolvedExecutable
	if s.resolver != nil {
		res, err := s.resolver.ResolveContractExecutable(ctx, contractID)
		if err == nil {
			resolved = res
		}
	}

	return &ContractInspection{
		ContractID: contractID,
		IsMember:   isMember,
		Member:     member,
		Resolved:   resolved,
	}, nil
}

func (s *Service) validateFleetID(id FleetID) error {
	if _, err := cap85.AddressToScAddress(id.Owner); err != nil {
		return fmt.Errorf("%w: %s", cap85.ErrInvalidContractID, id.Owner)
	}
	if _, err := cap85.ValidateAndDecodeTag(xdr.ScString(id.Tag)); err != nil {
		return fmt.Errorf("%w: %s", cap85.ErrInvalidTag, id.Tag)
	}
	return nil
}
