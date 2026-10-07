package cap85

import (
	"fmt"

	"github.com/stellar/go-stellar-sdk/strkey"
	"github.com/stellar/go-stellar-sdk/xdr"
)

// DecodeExternalExecutableRef decodes a ContractExecutable XDR union.
// Returns:
// - (ref, true, nil) for CONTRACT_EXECUTABLE_EXTERNAL_REF
// - (zero, false, nil) for CONTRACT_EXECUTABLE_WASM or CONTRACT_EXECUTABLE_STELLAR_ASSET
// - (zero, false, err) for malformed or unsupported executable types
func DecodeExternalExecutableRef(
	executable xdr.ContractExecutable,
) (ExternalExecutableRef, bool, error) {
	switch executable.Type {
	case xdr.ContractExecutableTypeContractExecutableWasm:
		return ExternalExecutableRef{}, false, nil

	case xdr.ContractExecutableTypeContractExecutableStellarAsset:
		return ExternalExecutableRef{}, false, nil

	case xdr.ContractExecutableTypeContractExecutableExternalRef:
		ref := executable.ExternalRef
		if ref == nil {
			return ExternalExecutableRef{}, false, fmt.Errorf("%w: missing external_ref payload in union", ErrInvalidExternalRef)
		}

		owner, err := DecodeScAddress(ref.ExecutableOwner)
		if err != nil {
			return ExternalExecutableRef{}, false, fmt.Errorf("%w: decode executable_owner: %v", ErrInvalidExternalRef, err)
		}

		tag, err := ValidateAndDecodeTag(ref.Tag)
		if err != nil {
			return ExternalExecutableRef{}, false, fmt.Errorf("%w: decode tag: %v", ErrInvalidExternalRef, err)
		}

		decoded := ExternalExecutableRef{
			Owner: owner,
			Tag:   tag,
		}
		if err := decoded.Validate(); err != nil {
			return ExternalExecutableRef{}, false, err
		}

		return decoded, true, nil

	default:
		return ExternalExecutableRef{}, false, fmt.Errorf("%w: unsupported contract executable type %d", ErrInvalidExternalRef, executable.Type)
	}
}

// DecodeScAddress converts an XDR ScAddress into its canonical StrKey string representation.
func DecodeScAddress(address xdr.ScAddress) (string, error) {
	str, err := address.String()
	if err != nil {
		return "", fmt.Errorf("encode ScAddress to strkey: %w", err)
	}
	if str == "" {
		return "", fmt.Errorf("empty ScAddress string representation")
	}
	return str, nil
}

// AddressToScAddress parses a StrKey address into an XDR ScAddress.
func AddressToScAddress(addrStr string) (xdr.ScAddress, error) {
	if addrStr == "" {
		return xdr.ScAddress{}, fmt.Errorf("%w: address cannot be empty", ErrInvalidContractID)
	}

	version, raw, err := strkey.DecodeAny(addrStr)
	if err != nil {
		return xdr.ScAddress{}, fmt.Errorf("%w: invalid strkey %q: %v", ErrInvalidContractID, addrStr, err)
	}

	switch version {
	case strkey.VersionByteContract:
		var contractID xdr.ContractId
		if len(raw) != len(contractID) {
			return xdr.ScAddress{}, fmt.Errorf("%w: invalid contract ID byte length %d", ErrInvalidContractID, len(raw))
		}
		copy(contractID[:], raw)
		return xdr.NewScAddress(xdr.ScAddressTypeScAddressTypeContract, contractID)

	case strkey.VersionByteAccountID:
		var pubkey xdr.Uint256
		if len(raw) != len(pubkey) {
			return xdr.ScAddress{}, fmt.Errorf("%w: invalid account ID byte length %d", ErrInvalidContractID, len(raw))
		}
		copy(pubkey[:], raw)
		accountID := xdr.AccountId{
			Type:    xdr.PublicKeyTypePublicKeyTypeEd25519,
			Ed25519: &pubkey,
		}
		return xdr.NewScAddress(xdr.ScAddressTypeScAddressTypeAccount, accountID)

	default:
		return xdr.ScAddress{}, fmt.Errorf("%w: unsupported strkey version byte %v", ErrInvalidContractID, version)
	}
}
