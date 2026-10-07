package cap85

import (
	"errors"
	"fmt"
)

var (
	// ErrNotExternalRef indicates that the contract executable is not a CAP-85 ExternalRef.
	ErrNotExternalRef = errors.New("contract executable is not an external reference")

	// ErrInvalidExternalRef indicates that the external reference is malformed or invalid.
	ErrInvalidExternalRef = errors.New("invalid external reference")

	// ErrBrokenReference indicates that the referenced owner or tag data cannot be resolved.
	ErrBrokenReference = errors.New("broken external reference")

	// ErrInvalidContractID indicates that the provided contract ID is not a valid Stellar StrKey address.
	ErrInvalidContractID = errors.New("invalid contract ID")

	// ErrInvalidTag indicates that the tag string is invalid.
	ErrInvalidTag = errors.New("invalid executable tag")
)

// FleetID uniquely identifies a fleet by owner address and tag.
type FleetID struct {
	Owner string
	Tag   string
}

// String returns a human-readable representation of the fleet identity.
func (f FleetID) String() string {
	return fmt.Sprintf("%s:%s", f.Owner, f.Tag)
}

// ExternalExecutableRef represents a CAP-85 external contract executable reference.
type ExternalExecutableRef struct {
	Owner string
	Tag   string
}

// FleetID returns the canonical FleetID for this external reference.
func (r ExternalExecutableRef) FleetID() FleetID {
	return FleetID{
		Owner: r.Owner,
		Tag:   r.Tag,
	}
}

// Validate checks that the reference contains valid, non-empty components.
func (r ExternalExecutableRef) Validate() error {
	if r.Owner == "" {
		return fmt.Errorf("%w: owner address cannot be empty", ErrInvalidExternalRef)
	}
	if r.Tag == "" {
		return fmt.Errorf("%w: tag cannot be empty", ErrInvalidTag)
	}
	return nil
}
