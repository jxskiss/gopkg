package slices

import "github.com/jxskiss/gopkg/v2/internal/constraints"

// Grow increases the slice's capacity, if necessary, to guarantee space for
// another n elements. After Grow(n), at least n elements can be appended
// to the slice without another allocation. If n is negative or too large to
// allocate the memory, Grow panics.
func Grow[S ~[]E, E any](slice S, n int) S {
	if cap(slice) >= len(slice)+n {
		return slice
	}
	out := make(S, len(slice), len(slice)+n)
	copy(out, slice)
	return out
}

// Clip removes unused capacity from the slice, returning s[:len(s):len(s)].
func Clip[S ~[]E, E any](s S) S {
	return s[:len(s):len(s)]
}

// Equal reports whether two slices are equal: the same length and all
// elements equal. If the lengths are different, Equal returns false.
// Otherwise, the elements are compared in increasing index order, and the
// comparison stops at the first unequal pair.
// Floating point NaNs are not considered equal.
func Equal[S ~[]E, E comparable](s1, s2 S) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i, v1 := range s1 {
		v2 := s2[i]
		if v1 != v2 {
			return false
		}
	}
	return true
}

// EqualFunc reports whether two slices are equal using a comparison
// function on each pair of elements. If the lengths are different,
// EqualFunc returns false. Otherwise, the elements are compared in
// increasing index order, and the comparison stops at the first index
// for which eq returns false.
func EqualFunc[E1, E2 any](s1 []E1, s2 []E2, eq func(E1, E2) bool) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i, v1 := range s1 {
		v2 := s2[i]
		if !eq(v1, v2) {
			return false
		}
	}
	return true
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

// IndexFunc returns the first index i satisfying f(s[i]),
// or -1 if none do.
func IndexFunc[S ~[]E, E any](s S, f func(E) bool) int {
	for i, v := range s {
		if f(v) {
			return i
		}
	}
	return -1
}

// LastIndex returns the index of the last occurrence of v in s,
// or -1 if not present.
func LastIndex[S ~[]E, E comparable](s S, v E) int {
	for i := len(s) - 1; i >= 0; i-- {
		if v == s[i] {
			return i
		}
	}
	return -1
}

// LastIndexFunc returns the last index i satisfying f(s[i]),
// or -1 if none do.
func LastIndexFunc[S ~[]E, E any](s S, f func(E) bool) int {
	for i := len(s) - 1; i >= 0; i-- {
		if f(s[i]) {
			return i
		}
	}
	return -1
}

// Contains reports whether v is present in s.
func Contains[S ~[]E, E comparable](s S, v E) bool {
	return Index(s, v) >= 0
}

// Insert inserts the values v... into s at index i,
// returning the modified slice.
// In the returned slice r, r[i] == v[0].
// Insert panics if i is out of range.
// Time complexity of this function is O(len(s) + len(v)).
func Insert[S ~[]E, E any](s S, i int, v ...E) S {
	tot := len(s) + len(v)
	if tot <= cap(s) {
		s2 := s[:tot]
		copy(s2[i+len(v):], s[i:])
		copy(s2[i:], v)
		return s2
	}
	s2 := make(S, tot)
	copy(s2, s[:i])
	copy(s2[i:], v)
	copy(s2[i+len(v):], s[i:])
	return s2
}

// Delete removes the elements s[i:j] from s, returning the modified slice.
// Delete panics if s[i:j] is not a valid slice of s.
// Delete modifies the contents of the slice s; it does not create a new slice.
// Delete is O(len(s)-(j-i)), so if many items must be deleted, it is better to
// make a single call deleting them all together than to delete one at a time.
func Delete[S ~[]E, E any](s S, i, j int) S {
	return append(s[:i], s[j:]...)
}

// Clone returns a copy of the slice.
// The elements are copied using assignment, so this is a shallow clone.
func Clone[S ~[]E, E any](s S) S {
	// Preserve nil in case it matters.
	if s == nil {
		return nil
	}
	return append(S([]E{}), s...)
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

// Diff returns a new slice containing the values which present
// in slice a, but not present in slice b.
func Diff[S ~[]E, E comparable](a, b S) S {
	if a == nil {
		return nil
	}
	bset := make(map[E]struct{}, len(b))
	for _, x := range b {
		bset[x] = struct{}{}
	}
	out := make(S, 0, max(0, len(a)-len(b)))
	for _, x := range a {
		if _, ok := bset[x]; !ok {
			out = append(out, x)
		}
	}
	return out
}

// Split splits a large slice []T to batches, it returns a slice
// of type [][]T whose elements are sub slices of slice.
func Split[S []E, E any](slice S, batch int) []S {
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

// Sum returns the sum value of the elements in the given slice.
func Sum[T constraints.Integer](slice []T) int64 {
	var sum int64
	for _, x := range slice {
		sum += int64(x)
	}
	return sum
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
