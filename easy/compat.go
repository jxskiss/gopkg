package easy

import (
	"net/http"

	"github.com/jxskiss/gopkg/v2/easy/ezhttp"
	"github.com/jxskiss/gopkg/v2/easy/ezmap"
	"github.com/jxskiss/gopkg/v2/internal/constraints"
)

// -------- slice utilities -------- //

// InInt32s tells whether the int32 value elem is in the slice.
func InInt32s(slice []int32, elem int32) bool {
	return Index(slice, elem) >= 0
}

// InInt64s tells whether the int64 value elem is in the slice.
func InInt64s(slice []int64, elem int64) bool {
	return Index(slice, elem) >= 0
}

// InStrings tells whether the string value elem is in the slice.
func InStrings(slice []string, elem string) bool {
	return Index(slice, elem) >= 0
}

// FilterInt32s iterates the given slice, it calls predicate(i) for i in
// range [0, n), where n is the length of the slice.
// It returns a new slice of elements for which predicate(i) returns true.
//
// Deprecated: the generic function Filter is favored over this.
func FilterInt32s(slice []int32, predicate func(i int) bool) []int32 {
	return Filter(func(i int, _ int32) bool { return predicate(i) }, slice)
}

// FilterInt64s iterates the given slice, it calls predicate(i) for i in
// range [0, n), where n is the length of the slice.
// It returns a new slice of elements for which predicate(i) returns true.
//
// Deprecated: the generic function Filter is favored over this.
func FilterInt64s(slice []int64, predicate func(i int) bool) []int64 {
	return Filter(func(i int, _ int64) bool { return predicate(i) }, slice)
}

// FilterStrings iterates the given slice, it calls predicate(i) for i in
// range [0, n), where n is the length of the slice.
// It returns a new slice of elements for which predicate(i) returns true.
//
// Deprecated: the generic function Filter is favored over this.
func FilterStrings(slice []string, predicate func(i int) bool) []string {
	return Filter(func(i int, _ string) bool { return predicate(i) }, slice)
}

// ReverseInt32s returns a new slice of the elements in reversed order.
//
// When inplace is true, the slice is reversed in place, it does not create
// a new slice, but returns the original slice with reversed order.
//
// Deprecated: the generic function Reverse is favored over this.
func ReverseInt32s(slice []int32, inplace bool) []int32 {
	return Reverse(slice, inplace)
}

// ReverseInt64s returns a new slice of the elements in reversed order.
//
// When inplace is true, the slice is reversed in place, it does not create
// a new slice, but returns the original slice with reversed order.
//
// Deprecated: the generic function Reverse is favored over this.
func ReverseInt64s(slice []int64, inplace bool) []int64 {
	return Reverse(slice, inplace)
}

// ReverseStrings returns a new slice of the elements in reversed order.
//
// When inplace is true, the slice is reversed in place, it does not create
// a new slice, but returns the original slice with reversed order.
//
// Deprecated: the generic function Reverse is favored over this.
func ReverseStrings(slice []string, inplace bool) []string {
	return Reverse(slice, inplace)
}

// SplitSlice splits a large slice []T to batches, it returns a slice
// of type [][]T whose elements are sub slices of slice.
//
// Deprecated: the generic function Split is favored over this.
func SplitSlice[S ~[]E, E any](slice S, batch int) interface{} {
	return Split[S, E](slice, batch)
}

// SplitInt32s splits a large int32 slice to batches.
func SplitInt32s(slice []int32, batch int) [][]int32 {
	return Split(slice, batch)
}

// SplitInt64s splits a large int64 slice to batches.
func SplitInt64s(slice []int64, batch int) [][]int64 {
	return Split(slice, batch)
}

// SplitStrings splits a large string slice to batches.
func SplitStrings(slice []string, batch int) [][]string {
	return Split(slice, batch)
}

// UniqueInt32s returns a new slice containing the elements of the given
// slice in same order, but filter out duplicate values.
//
// When inplace is true, it does not create a new slice, the unique values
// will be written to the input slice from the beginning.
//
// Deprecated: the generic function Unique is favored over this.
func UniqueInt32s(slice []int32, inplace bool) []int32 {
	return Unique(slice, inplace)
}

// UniqueInt64s returns a new slice containing the elements of the given
// slice in same order, but filter out duplicate values.
//
// When inplace is true, it does not create a new slice, the unique values
// will be written to the input slice from the beginning.
//
// Deprecated: the generic function Unique is favored over this.
func UniqueInt64s(slice []int64, inplace bool) []int64 {
	return Unique(slice, inplace)
}

// UniqueStrings returns a new slice containing the elements of the given
// slice in same order, but filter out duplicate values.
//
// When inplace is true, it does not create a new slice, the unique values
// will be written to the input slice from the beginning.
//
// Deprecated: the generic function Unique is favored over this.
func UniqueStrings(slice []string, inplace bool) []string {
	return Unique(slice, inplace)
}

// DiffInt32s returns a new int32 slice containing the values which present
// in slice a but not present in slice b.
//
// Deprecated: the generic function Diff is favored over this.
func DiffInt32s(a []int32, b []int32) []int32 {
	return Diff(a, b)
}

// DiffInt64s returns a new int64 slice containing the values which present
// in slice a but not present in slice b.
//
// Deprecated: the generic function Diff is favored over this.
func DiffInt64s(a []int64, b []int64) []int64 {
	return Diff(a, b)
}

// DiffStrings returns a new string slice containing the values which
// present in slice a but not present in slice b.
//
// Deprecated: the generic function Diff is favored over this.
func DiffStrings(a []string, b []string) []string {
	return Diff(a, b)
}

// -------- map utilities -------- //

// MapKeys returns the keys of the map m.
// The keys will be in an indeterminate order.
//
// Deprecated: the generic function Keys is favored over this.
func MapKeys[M ~map[K]V, K comparable, V any](m M) interface{} {
	return Keys[M, K, V](m)
}

// MapValues returns the values of the map m.
// The values will be in an indeterminate order.
//
// Deprecated: the generic function Values is favored over this.
func MapValues[M ~map[K]V, K comparable, V any](m M) interface{} {
	return Values[M, K, V](m)
}

// IntKeys returns a int64 slice containing all the keys present
// in the map, in an indeterminate order.
//
// Deprecated: the generic function Keys is favored over this.
func IntKeys[M ~map[K]V, K constraints.Integer, V any](m M) (keys []int64) {
	keys = make([]int64, 0, len(m))
	for k := range m {
		keys = append(keys, int64(k))
	}
	return
}

// IntValues returns a int64 slice containing all the values present
// in the map, in an indeterminate order.
//
// Deprecated: the generic function Values is favored over this.
func IntValues[M ~map[K]V, K comparable, V constraints.Integer](m M) (values []int64) {
	values = make([]int64, 0, len(m))
	for _, v := range m {
		values = append(values, int64(v))
	}
	return
}

// StringKeys returns a string slice containing all the keys present
// in the map, in an indeterminate order.
//
// Deprecated: the generic function Keys is favored over this.
func StringKeys[M ~map[K]V, K ~string, V any](m M) (keys []string) {
	keys = make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, string(k))
	}
	return
}

// StringValues returns a string slice containing all the values present
// in the map, in an indeterminate order.
//
// Deprecated: the generic function Values is favored over this.
func StringValues[M ~map[K]V, K comparable, V ~string](m M) (values []string) {
	values = make([]string, 0, len(m))
	for _, v := range m {
		values = append(values, string(v))
	}
	return
}

// -------- ezmap alias names -------- //

// Map is an alias name of ezmap.Map.
//
// Deprecated: please use ezmap.Map directly, this alias name will be
// removed in future releases.
type Map = ezmap.Map

// SafeMap is an alias name of ezmap.SafeMap.
//
// Deprecated: please use ezmap.SafeMap directly, this alias name will
// be removed in future releases.
type SafeMap = ezmap.SafeMap

// NewMap is an alias name of ezmap.NewMap.
//
// Deprecated: please use ezmap.NewMap directly, this alias name will
// be removed in future releases.
func NewMap() Map { return ezmap.NewMap() }

// NewSafeMap is an alias name of ezmap.NewSafeMap.
//
// Deprecated: please use ezmap.NewSafeMap directly, this alias name will
// be removed in future releases.
func NewSafeMap() *SafeMap { return ezmap.NewSafeMap() }

// -------- http utilities -------- //

// Request represents a request and options to send with the Do function.
//
// Deprecated: moved to package httputil, please use ezhttp.Request
// instead of this.
type Request = ezhttp.Request

// DoRequest is a convenient function to send request and control redirect
// and debug options.
//
// Deprecated: this function has been moved to ezhttp.Do, please use
// ezhttp.Do instead of this.
func DoRequest(req *Request) (header http.Header, respContent []byte, status int, err error) {
	return ezhttp.Do(req)
}
