package ptr

import (
	"strconv"

	"github.com/jxskiss/gopkg/v2/internal/constraints"
)

// IntToStringp converts Integer x to a string pointer.
func IntToStringp[T constraints.Integer](x T) *string {
	str := strconv.FormatInt(int64(x), 10)
	return &str
}

// IntpToStringp converts x to a string pointer.
// It returns nil if x is nil.
func IntpToStringp[T constraints.Integer](x *T) *string {
	if x == nil {
		return nil
	}
	str := strconv.FormatInt(int64(*x), 10)
	return &str
}

// IntpToString converts x to a string.
// It returns an empty string if x is nil.
func IntpToString[T constraints.Integer](x *T) string {
	if x == nil {
		return ""
	}
	return strconv.FormatInt(int64(*x), 10)
}
