package cap85

import (
	"errors"
	"testing"

	"github.com/stellar/go-stellar-sdk/xdr"
)

func TestTagPreservation(t *testing.T) {
	cases := []string{
		"vault-v1",
		"  vault-with-leading-spaces",
		"vault-with-trailing-spaces  ",
		"Vault-MixedCase-V2",
		"tag_with_underscores_and_symbols!#$",
		"tag:with:colons",
		"tag/with/slashes",
		"tag-with-utf8-🚀",
	}

	for _, tc := range cases {
		decoded, err := ValidateAndDecodeTag(xdr.ScString(tc))
		if err != nil {
			t.Fatalf("unexpected error decoding tag %q: %v", tc, err)
		}
		if decoded != tc {
			t.Errorf("tag was altered: expected %q, got %q", tc, decoded)
		}

		// Verify BuildTagScVal builds the correct SCV_EXECUTABLE_TAG ScVal
		scVal, err := BuildTagScVal(tc)
		if err != nil {
			t.Fatalf("unexpected error building ScVal for tag %q: %v", tc, err)
		}
		if scVal.Type != xdr.ScValTypeScvExecutableTag {
			t.Errorf("expected ScValType %v, got %v", xdr.ScValTypeScvExecutableTag, scVal.Type)
		}
		valTag := string(*scVal.ExecutableTag)
		if valTag != tc {
			t.Errorf("built ScVal tag altered: expected %q, got %q", tc, valTag)
		}
	}
}

func TestTagValidationRejections(t *testing.T) {
	// Empty tag
	_, err := ValidateAndDecodeTag(xdr.ScString(""))
	if !errors.Is(err, ErrInvalidTag) {
		t.Errorf("expected ErrInvalidTag for empty tag, got %v", err)
	}

	_, err = BuildTagScVal("")
	if !errors.Is(err, ErrInvalidTag) {
		t.Errorf("expected ErrInvalidTag for empty tag ScVal, got %v", err)
	}

	// Invalid UTF-8
	invalidUTF8 := string([]byte{0xff, 0xfe, 0xfd})
	_, err = ValidateAndDecodeTag(xdr.ScString(invalidUTF8))
	if !errors.Is(err, ErrInvalidTag) {
		t.Errorf("expected ErrInvalidTag for invalid UTF-8, got %v", err)
	}

	_, err = BuildTagScVal(invalidUTF8)
	if !errors.Is(err, ErrInvalidTag) {
		t.Errorf("expected ErrInvalidTag for invalid UTF-8 ScVal, got %v", err)
	}
}
