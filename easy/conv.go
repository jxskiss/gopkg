package easy

import (
	"strconv"

	"github.com/jxskiss/gopkg/v2/internal/constraints"
)

// ConvInts converts slice of integers of type T1 to a new
// slice of integers of type T2.
// The input data must be convertable to T2.
func ConvInts[T1, T2 constraints.Integer](slice []T1) []T2 {
	if len(slice) == 0 {
		return nil
	}
	out := make([]T2, len(slice))
	for i, x := range slice {
		out[i] = T2(x)
	}
	return out
}

// FormatInts converts slice of integers to a slice of strings.
// It returns nil if there is no element in given slice.
func FormatInts[T constraints.Integer](slice []T, base int) []string {
	if len(slice) == 0 {
		return nil
	}
	out := make([]string, len(slice))
	for i, x := range slice {
		out[i] = strconv.FormatInt(int64(x), base)
	}
	return out
}

// ParseInts converts slice of strings to a slice of integers.
// It returns nil if there is no element in given slice.
//
// Note that if the input data contains non-integer strings, the
// errors returned from strconv.ParseInt are ignored, and the
// returned slice will have less elements than the input slice.
func ParseInts[T constraints.Integer](slice []string, base int) []T {
	if len(slice) == 0 {
		return nil
	}
	out := make([]T, 0, len(slice))
	for _, x := range slice {
		iv, err := strconv.ParseInt(x, base, 0)
		if err == nil {
			out = append(out, T(iv))
		}
	}
	return out
}

// ToMap converts the given slice to a map, using elements from the
// slice as keys and true as values.
func ToMap[S ~[]E, E comparable](slice S) map[E]bool {
	if len(slice) == 0 {
		return nil
	}
	out := make(map[E]bool, len(slice))
	for _, elem := range slice {
		out[elem] = true
	}
	return out
}

// ToInterfaceSlice returns a []interface{} containing elements from slice.
func ToInterfaceSlice[S ~[]E, E any](slice S) []interface{} {
	if len(slice) == 0 {
		return nil
	}
	out := make([]interface{}, len(slice))
	for i, elem := range slice {
		out[i] = elem
	}
	return out
}
