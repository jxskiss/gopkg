package easy

import "github.com/jxskiss/gopkg/v2/internal/constraints"

// Diff returns a new slice containing the values which present
// in slice, but not present in others.
func Diff[S ~[]E, E comparable](slice S, others ...S) S {
	if len(slice) == 0 {
		return nil
	}
	s2set := make(map[E]struct{})
	for _, s := range others {
		for _, x := range s {
			s2set[x] = struct{}{}
		}
	}
	out := make(S, 0, len(slice))
	for _, x := range slice {
		if _, ok := s2set[x]; !ok {
			out = append(out, x)
		}
	}
	return out
}

// Filter iterates the given slices, it calls predicate(i, elem) for
// each elem in the slices and returns a new slice of elements for which
// predicate(i, elem) returns true.
func Filter[S ~[]E, E any](predicate func(i int, elem E) bool, slices ...S) S {
	if len(slices) == 0 {
		return nil
	}
	out := make(S, 0, len(slices[0]))
	for _, s := range slices {
		for i, e := range s {
			if predicate(i, e) {
				out = append(out, e)
			}
		}
	}
	return out
}

// IndexFunc iterates the given slice, it calls predicate(i) for i in
// range [0, n) where n is the length of the slice.
// When predicate(i) returns true, it stops and returns the index i.
func IndexFunc[E any](slice []E, predicate func(i int) bool) int {
	for i := range slice {
		if predicate(i) {
			return i
		}
	}
	return -1
}

// LastIndexFunc iterates the given slice, it calls predicate(i) for i in
// range [0, n) in descending order, where n is the length of the slice.
// When predicate(i) returns true, it stops and returns the index i.
func LastIndexFunc[E any](slice []E, predicate func(i int) bool) int {
	for i := len(slice) - 1; i >= 0; i-- {
		if predicate(i) {
			return i
		}
	}
	return -1
}

// IJ represents a batch index of i, j.
type IJ struct{ I, J int }

// SplitBatch splits a large number to batch, it's mainly designed to
// help operations with large slice, such as inserting lots of records
// into database, or logging lots of identifiers, etc.
func SplitBatch(total, batch int) []IJ {
	if total <= 0 {
		return nil
	}
	if batch <= 0 {
		return []IJ{{0, total}}
	}
	n := total/batch + 1
	ret := make([]IJ, n)
	idx := 0
	for i, j := 0, batch; idx < n && i < total; i, j = i+batch, j+batch {
		if j > total {
			j = total
		}
		ret[idx] = IJ{i, j}
		idx++
	}
	return ret[:idx]
}

// Split splits a large slice []T to batches, it returns a slice
// of type [][]T whose elements are sub slices of slice.
func Split[S ~[]E, E any](slice S, batch int) []S {
	if len(slice) == 0 {
		return nil
	}
	if batch <= 0 {
		return []S{slice}
	}
	n := len(slice) / batch
	ret := make([]S, 0, n+1)
	for i := 0; i < n*batch; i += batch {
		ret = append(ret, slice[i:i+batch])
	}
	if last := n * batch; last < len(slice) {
		ret = append(ret, slice[last:])
	}
	return ret
}

// Reverse returns a slice of the elements in reversed order.
// When inplace is true, it does not allocate new memory, but the slice
// is reversed in place.
func Reverse[S ~[]E, E any](s S, inplace bool) S {
	if s == nil {
		return nil
	}
	out := s
	if !inplace {
		out = make(S, len(s))
		copy(out, s)
	}
	i, j := 0, len(s)-1
	for i < j {
		out[i], out[j] = out[j], out[i]
		i++
		j--
	}
	return out
}

// Unique returns a slice containing the elements of the given
// slice in same order, but removes duplicate values.
// When inplace is true, it does not allocate new memory, the unique values
// will be written to the input slice from the beginning.
func Unique[S ~[]E, E comparable](s S, inplace bool) S {
	if s == nil {
		return nil
	}
	seen := make(map[E]struct{})
	out := s[:0]
	if !inplace {
		out = make(S, 0)
	}
	for _, x := range s {
		if _, ok := seen[x]; !ok {
			seen[x] = struct{}{}
			out = append(out, x)
		}
	}
	return out
}

// Sum returns the sum value of the elements in the given slice.
func Sum[T constraints.Integer](slice []T) int64 {
	var sum int64
	for _, x := range slice {
		sum += int64(x)
	}
	return sum
}
