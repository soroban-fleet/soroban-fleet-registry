package cap85

import (
	"errors"
	"testing"
)

func TestFleetIDString(t *testing.T) {
	id := FleetID{
		Owner: "CAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		Tag:   "vault-v1",
	}
	expected := "CAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA:vault-v1"
	if id.String() != expected {
		t.Fatalf("expected %s, got %s", expected, id.String())
	}
}

func TestExternalExecutableRefValidation(t *testing.T) {
	ref := ExternalExecutableRef{
		Owner: "CAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		Tag:   "vault-v1",
	}
	if err := ref.Validate(); err != nil {
		t.Fatalf("expected valid ref, got err: %v", err)
	}

	fleetID := ref.FleetID()
	if fleetID.Owner != ref.Owner || fleetID.Tag != ref.Tag {
		t.Fatalf("mismatched fleet ID: got %+v", fleetID)
	}

	emptyOwner := ExternalExecutableRef{Owner: "", Tag: "vault-v1"}
	if err := emptyOwner.Validate(); !errors.Is(err, ErrInvalidExternalRef) {
		t.Fatalf("expected ErrInvalidExternalRef, got %v", err)
	}

	emptyTag := ExternalExecutableRef{Owner: "CAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA", Tag: ""}
	if err := emptyTag.Validate(); !errors.Is(err, ErrInvalidTag) {
		t.Fatalf("expected ErrInvalidTag, got %v", err)
	}
}
