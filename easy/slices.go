package easy

import (
	"sort"

	"github.com/jxskiss/gopkg/v2/internal/constraints"
)

// Clip removes unused capacity from the slice, returning s[:len(s):len(s)].
func Clip[S ~[]E, E any](s S) S {
	return s[:len(s):len(s)]
}

// Copy copies a slice to be a new one.
// optionalCap optionally specifies the capacity of the new slice.
func Copy[S ~[]E, E any](s S, optionalCap ...int) S {
	copyCap := len(s)
	if len(optionalCap) > 0 && optionalCap[0] > copyCap {
		copyCap = optionalCap[0]
	}
	out := make(S, len(s), copyCap)
	copy(out, s)
	return out
}

// Concat concatenates given slices into a single slice.
func Concat[S ~[]E, E any](slices ...S) S {
	n := 0
	for _, s := range slices {
		n += len(s)
	}
	out := make(S, 0, n)
	for _, s := range slices {
		out = append(out, s...)
	}
	return out
}

// Count iterates slices, it calls predicate(elem) for
// each elem in the slices and returns the count of elements for which
// predicate(elem) returns true.
func Count[S ~[]E, E any](predicate func(elem E) bool, slices ...S) int {
	count := 0
	for _, s := range slices {
		for _, e := range s {
			if predicate(e) {
				count++
			}
		}
	}
	return count
}

// Diff allocates and returns a new slice which contains the values
// which present in slice, but not present in others.
//
// If length of slice is zero, it returns nil.
func Diff[S ~[]E, E comparable](slice S, others ...S) S {
	return diffSlice(false, slice, others...)
}

// DiffInplace returns a slice which contains the values which present
// in slice, but not present in others.
// It does not allocate new memory, but modifies slice in-place.
//
// If length of slice is zero, it returns nil.
func DiffInplace[S ~[]E, E comparable](slice S, others ...S) S {
	return diffSlice(true, slice, others...)
}

func diffSlice[S ~[]E, E comparable](inplace bool, slice S, others ...S) S {
	if len(slice) == 0 {
		return nil
	}
	s2set := make(map[E]struct{})
	for _, s := range others {
		for _, x := range s {
			s2set[x] = struct{}{}
		}
	}
	out := slice[:0]
	if !inplace {
		out = make(S, 0, len(slice))
	}
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

// InSlice tells whether the value elem is in the slice.
func InSlice[E comparable](slice []E, elem E) bool {
	return Index(slice, elem) >= 0
}

// Index returns the index of the first occurrence of v in s,
// or -1 if not present.
func Index[S ~[]E, E comparable](s S, v E) int {
	for i, vs := range s {
		if v == vs {
			return i
		}
	}
	return -1
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

// LastIndex returns the index of the last instance of v in s,
// or -1 if v is not present in s.
func LastIndex[S ~[]E, E comparable](s S, v E) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == v {
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

// Repeat returns a new slice consisting of count copies of the slice s.
//
// It panics if count is zero or negative or if
// the result of (len(s) * count) overflows.
func Repeat[S ~[]E, E any](s S, count int) S {
	if count <= 0 {
		panic("zero or negative Repeat count")
	} else if len(s)*count/count != len(s) {
		panic("Repeat count causes overflow")
	}

	out := make(S, 0, count*len(s))
	for i := 0; i < count; i++ {
		out = append(out, s...)
	}
	return out
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
//
// Given different input, the duplication rate may be varying,
// for large input, this function does not assume any specific workload type,
// it allocates initial memory of size `len(s)/2`, thus for slice that
// most elements are duplicate, it allocates memory more than need,
// but for slice that no value is duplicate, it triggers memory allocation
// more than once.
// For large slice in performance critical use-case, user is recommended to
// write a custom function that is fine-tuned for specific workload to get
// the best performance.
func Unique[S ~[]E, E comparable](s S, inplace bool) S {
	if s == nil {
		return nil
	}
	var out S
	if inplace {
		out = s[:0]
	}

	// According to benchmark results, 128 is a reasonable choice
	// to balance the performance of different algorithms and the cost
	// of memory allocation.
	// See BenchmarkUnique* in slices_test.go.
	if len(s) <= 128 {
		return uniqueByLoopCmp(out, s)
	}
	return uniqueByHashset(out, s)
}

func uniqueByLoopCmp[S ~[]E, E comparable](dst, src S) S {
	if cap(dst) == 0 {
		dst = make(S, 0, len(src))
	}
	for _, x := range src {
		isDup := false
		for i := range dst {
			if x == dst[i] {
				isDup = true
				break
			}
		}
		if !isDup {
			dst = append(dst, x)
		}
	}
	return dst
}

func uniqueByHashset[S ~[]E, E comparable](dst, src S) S {
	if cap(dst) == 0 {
		dst = make(S, 0, len(src)/2)
	}
	seen := make(map[E]struct{}, len(src)/2)
	for _, x := range src {
		if _, ok := seen[x]; !ok {
			seen[x] = struct{}{}
			dst = append(dst, x)
		}
	}
	return dst
}

// UniqueFunc returns a slice containing the elements of the given slice
// in same order, but removes deduplicate values, it calls f for each
// element and uses the returned value to check duplication.
// When inplace is true, it does not allocate new memory, the unique values
// will be written to the input slice from the beginning.
//
// Given different input, the duplication rate may be varying,
// for large input, this function does not assume any specific workload type,
// it allocates initial memory of size `len(s)/2`, thus for slice that
// most elements are duplicate, it allocates memory more than need,
// but for slice that no value is duplicate, it triggers memory allocation
// more than once.
// For large slice in performance critical use-case, user is recommended to
// write a custom function that is fine-tuned for specific workload to get
// the best performance.
func UniqueFunc[S ~[]E, E any, C comparable](s S, inplace bool, f func(E) C) S {
	if s == nil {
		return nil
	}
	var out S
	if inplace {
		out = s[:0]
	}
	if len(s) <= 128 {
		return uniqueFuncByLoopCmp(out, s, f)
	}
	return uniqueFuncByHashset(out, s, f)
}

func uniqueFuncByLoopCmp[S ~[]E, E any, C comparable](dst, src S, f func(E) C) S {
	if cap(dst) == 0 {
		dst = make(S, 0, len(src))
	}
	seen := make([]C, 0, len(src))
	for _, x := range src {
		c := f(x)
		isDup := false
		for i := range seen {
			if c == seen[i] {
				isDup = true
				break
			}
		}
		if !isDup {
			seen = append(seen, c)
			dst = append(dst, x)
		}
	}
	return dst
}

func uniqueFuncByHashset[S ~[]E, E any, C comparable](dst, src S, f func(E) C) S {
	if cap(dst) == 0 {
		dst = make(S, 0, len(src)/2)
	}
	seen := make(map[C]struct{}, len(src)/2)
	for _, x := range src {
		c := f(x)
		if _, ok := seen[c]; !ok {
			seen[c] = struct{}{}
			dst = append(dst, x)
		}
	}
	return dst
}

// Sum returns the sum value of the elements in the given slice.
func Sum[T constraints.Integer](slice []T) int64 {
	var sum int64
	for _, x := range slice {
		sum += int64(x)
	}
	return sum
}

// SumFloat returns the sum value of the elements in the given slice,
// as a float64 value.
func SumFloat[T constraints.RealNumber](slice []T) float64 {
	var sum float64
	for _, x := range slice {
		sum += float64(x)
	}
	return sum
}

// Sort sorts the given slice ascending and returns it.
func Sort[S ~[]E, E constraints.Ordered](s S) S {
	sort.Slice(s, func(i, j int) bool {
		return s[i] < s[j]
	})
	return s
}

// SortDesc sorts the given slice descending and returns it.
func SortDesc[S ~[]E, E constraints.Ordered](s S) S {
	sort.Slice(s, func(i, j int) bool {
		return s[j] < s[i]
	})
	return s
}
