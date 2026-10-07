package cap85

import (
	"fmt"
	"unicode/utf8"

	"github.com/stellar/go-stellar-sdk/xdr"
)

// ValidateAndDecodeTag validates a CAP-85 ScString tag and returns its exact canonical string.
// Per CAP-85 specification:
// - Case must NOT be normalized
// - Whitespace must NOT be trimmed
// - Unicode normalization must NOT be invented
// - Invalid UTF-8 or empty tags must NOT be silently recovered
func ValidateAndDecodeTag(tag xdr.ScString) (string, error) {
	tagStr := string(tag)
	if len(tagStr) == 0 {
		return "", fmt.Errorf("%w: tag cannot be empty", ErrInvalidTag)
	}
	if !utf8.ValidString(tagStr) {
		return "", fmt.Errorf("%w: tag contains invalid UTF-8 bytes", ErrInvalidTag)
	}
	return tagStr, nil
}

// BuildTagScVal creates an ScVal of type SCV_EXECUTABLE_TAG containing the given tag string.
func BuildTagScVal(tag string) (xdr.ScVal, error) {
	if len(tag) == 0 {
		return xdr.ScVal{}, fmt.Errorf("%w: tag cannot be empty", ErrInvalidTag)
	}
	if !utf8.ValidString(tag) {
		return xdr.ScVal{}, fmt.Errorf("%w: tag contains invalid UTF-8 bytes", ErrInvalidTag)
	}
	scTag := xdr.ScString(tag)
	val, err := xdr.NewScVal(xdr.ScValTypeScvExecutableTag, scTag)
	if err != nil {
		return xdr.ScVal{}, fmt.Errorf("create executable tag ScVal: %w", err)
	}
	return val, nil
}
