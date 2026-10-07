package fleet

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/soroban-fleet/soroban-fleet-registry/internal/cap85"
)

type mockResolver struct {
	resolved map[string]cap85.ResolvedExecutable
}

func (m *mockResolver) ResolveContractExecutable(ctx context.Context, contractID string) (cap85.ResolvedExecutable, error) {
	if res, ok := m.resolved[contractID]; ok {
		return res, nil
	}
	return cap85.ResolvedExecutable{}, cap85.ErrBrokenReference
}

func (m *mockResolver) ResolveContractExecutablesBatch(ctx context.Context, contractIDs []string) (map[string]cap85.ResolvedExecutable, error) {
	res := make(map[string]cap85.ResolvedExecutable)
	for _, cid := range contractIDs {
		if r, ok := m.resolved[cid]; ok {
			res[cid] = r
		}
	}
	return res, nil
}

func TestFleetService_Validation(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	repo := NewPostgresRepository(db)
	svc := NewService(repo, nil)
	ctx := context.Background()

	// Invalid owner
	_, err := svc.GetFleet(ctx, FleetID{Owner: "invalid", Tag: "v1"})
	if !errors.Is(err, cap85.ErrInvalidContractID) {
		t.Errorf("expected ErrInvalidContractID, got %v", err)
	}

	// Invalid tag
	validOwner := makeContractID(10)
	_, err = svc.GetFleet(ctx, FleetID{Owner: validOwner, Tag: ""})
	if !errors.Is(err, cap85.ErrInvalidTag) {
		t.Errorf("expected ErrInvalidTag, got %v", err)
	}
}

func TestFleetService_InspectContract(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	_, _ = db.Exec("TRUNCATE TABLE fleet_members, fleets CASCADE")

	repo := NewPostgresRepository(db)

	ownerID := makeContractID(11)
	fleetID := FleetID{Owner: ownerID, Tag: "vault-v1"}
	_ = repo.UpsertFleet(context.Background(), &Fleet{
		ID:                fleetID,
		FirstSeenLedger:   10,
		LastSeenLedger:    10,
		LastIndexedLedger: 10,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	})

	cid := makeContractID(12)
	member := &FleetMember{
		ContractID:      cid,
		FleetID:         fleetID,
		WASMHash:        []byte{1, 2, 3},
		FirstSeenLedger: 10,
		LastSeenLedger:  10,
		Active:          true,
	}
	_ = repo.UpsertMember(context.Background(), member)

	mockRes := &mockResolver{
		resolved: map[string]cap85.ResolvedExecutable{
			cid: {
				Kind:     "EXTERNAL_REF",
				Fleet:    &fleetID,
				WASMHash: []byte{1, 2, 3},
			},
		},
	}

	svc := NewService(repo, mockRes)
	insp, err := svc.InspectContract(context.Background(), cid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !insp.IsMember {
		t.Errorf("expected isMember=true")
	}
	if insp.Member == nil || insp.Member.ContractID != cid {
		t.Errorf("expected member %s, got %+v", cid, insp.Member)
	}
	if insp.Resolved.Kind != "EXTERNAL_REF" {
		t.Errorf("expected resolved kind EXTERNAL_REF, got %s", insp.Resolved.Kind)
	}
}

func TestMembershipManager_Lifecycle(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	_, _ = db.Exec("TRUNCATE TABLE fleet_members, fleets CASCADE")

	repo := NewPostgresRepository(db)
	mm := NewMembershipManager(repo)
	ctx := context.Background()

	owner1 := makeContractID(20)
	fleet1 := FleetID{Owner: owner1, Tag: "tag-1"}
	fleet2 := FleetID{Owner: owner1, Tag: "tag-2"}

	_ = repo.UpsertFleet(ctx, &Fleet{ID: fleet1, FirstSeenLedger: 1, LastSeenLedger: 1, LastIndexedLedger: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()})
	_ = repo.UpsertFleet(ctx, &Fleet{ID: fleet2, FirstSeenLedger: 1, LastSeenLedger: 1, LastIndexedLedger: 1, CreatedAt: time.Now(), UpdatedAt: time.Now()})

	cid := makeContractID(21)

	// 1. Register in fleet1
	tx, _ := repo.BeginTx(ctx)
	if err := mm.RegisterOrUpdateMemberTx(ctx, tx, cid, fleet1, []byte{1}, 100); err != nil {
		_ = tx.Rollback()
		t.Fatalf("register member: %v", err)
	}
	_ = tx.Commit()

	f1, _ := repo.GetFleet(ctx, fleet1)
	if f1.MemberCount != 1 {
		t.Errorf("expected fleet1 member count 1, got %d", f1.MemberCount)
	}

	// 2. Move contract to fleet2
	tx, _ = repo.BeginTx(ctx)
	if err := mm.RegisterOrUpdateMemberTx(ctx, tx, cid, fleet2, []byte{2}, 200); err != nil {
		_ = tx.Rollback()
		t.Fatalf("move member: %v", err)
	}
	_ = tx.Commit()

	f1, _ = repo.GetFleet(ctx, fleet1)
	f2, _ := repo.GetFleet(ctx, fleet2)
	if f1.MemberCount != 0 {
		t.Errorf("expected fleet1 member count 0 after move, got %d", f1.MemberCount)
	}
	if f2.MemberCount != 1 {
		t.Errorf("expected fleet2 member count 1 after move, got %d", f2.MemberCount)
	}

	// 3. Deactivate from fleet2
	tx, _ = repo.BeginTx(ctx)
	if err := mm.DeactivateMemberTx(ctx, tx, cid, 300); err != nil {
		_ = tx.Rollback()
		t.Fatalf("deactivate: %v", err)
	}
	_ = tx.Commit()

	f2, _ = repo.GetFleet(ctx, fleet2)
	if f2.MemberCount != 0 {
		t.Errorf("expected fleet2 member count 0 after deactivation, got %d", f2.MemberCount)
	}

	mem, _ := repo.GetMember(ctx, cid)
	if mem.Active {
		t.Errorf("expected member to be inactive")
	}
}
