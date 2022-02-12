package easy

import (
	"strconv"

	"github.com/jxskiss/gopkg/v2/internal/constraints"
)

// FormatInts converts slice of integers to a slice of strings.
// It returns nil if there is no element in given slice.
func FormatInts[T constraints.Integer](slice []T) []string {
	if len(slice) == 0 {
		return nil
	}
	out := make([]string, 0, len(slice))
	for _, x := range slice {
		str := strconv.FormatInt(int64(x), 10)
		out = append(out, str)
	}
	return out
}

// ParseInts converts slice of strings to a slice of integers.
// It returns nil if there is no element in given slice.
func ParseInts[T constraints.Integer](slice []string) []T {
	if len(slice) == 0 {
		return nil
	}
	out := make([]T, 0, len(slice))
	for _, x := range slice {
		iv, err := strconv.ParseInt(x, 0, 0)
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

// ToSlice returns a slice consisting keys from the given map
// whose value are true.
//
// If you need all keys in the map without checking the values, you should
// use maps.Keys(m).
func ToSlice[M ~map[K]bool, K comparable](m M) []K {
	if len(m) == 0 {
		return nil
	}
	out := make([]K, 0, len(m))
	for k, val := range m {
		if val {
			out = append(out, k)
		}
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
