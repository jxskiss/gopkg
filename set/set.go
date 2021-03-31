package set

//go:generate go run template.go

import (
	"fmt"
	"reflect"
)

const minSize = 8

// Set is set collection of general type.
// The zero value of Set is an empty instance ready to use. A zero Set
// value shall not be copied, or it may result incorrect behavior.
type Set struct {
	m map[interface{}]struct{}
}

// NewSet creates a Set instance and add the given values into the set.
// If given only one param which is a slice, the elements of the slice
// will be added into the set using reflection.
func NewSet(vals ...interface{}) Set {
	size := max(len(vals), minSize)
	set := Set{
		m: make(map[interface{}]struct{}, size),
	}
	if len(vals) == 1 && reflect.TypeOf(vals[0]).Kind() == reflect.Slice {
		values := reflect.ValueOf(vals[0])
		for i := 0; i < values.Len(); i++ {
			set.m[values.Index(i).Interface()] = struct{}{}
		}
	} else {
		set.Add(vals...)
	}
	return set
}

// NewSetWithSize creates Set instance with given initial size.
func NewSetWithSize(size int) Set {
	set := Set{
		m: make(map[interface{}]struct{}, size),
	}
	return set
}

// Size returns the size of the set.
func (s *Set) Size() int { return len(s.m) }

// Add adds the given values into the set.
// If given only on param and which is a slice, the elements of the slice
// will be added into the set using reflection.
func (s *Set) Add(vals ...interface{}) {
	if s.m == nil {
		size := max(len(vals), minSize)
		s.m = make(map[interface{}]struct{}, size)
	}
	if len(vals) == 1 && reflect.TypeOf(vals[0]).Kind() == reflect.Slice {
		values := reflect.ValueOf(vals[0])
		for i := 0; i < values.Len(); i++ {
			s.m[values.Index(i).Interface()] = struct{}{}
		}
		return
	}

	for idx := range vals {
		s.m[vals[idx]] = struct{}{}
	}
}

// Del deletes values from the set.
func (s *Set) Del(vals ...interface{}) {
	if len(vals) == 1 && reflect.TypeOf(vals[0]).Kind() == reflect.Slice {
		values := reflect.ValueOf(vals[0])
		for i := 0; i < values.Len(); i++ {
			delete(s.m, values.Index(i).Interface())
		}
		return
	}

	for idx := range vals {
		delete(s.m, vals[idx])
	}
}

// Pop pops an element from the set, in no particular order.
func (s *Set) Pop() interface{} {
	for val := range s.m {
		delete(s.m, val)
		return val
	}
	return nil
}

// Iterate iterates the set in no particular order and call the given
// function for each set element.
func (s *Set) Iterate(fn func(interface{})) {
	for val := range s.m {
		fn(val)
	}
}

// Contains returns true if the set contains all the values.
func (s *Set) Contains(vals ...interface{}) bool {
	if len(vals) == 0 {
		return false
	}
	for _, v := range vals {
		if _, ok := s.m[v]; !ok {
			return false
		}
	}
	return true
}

// ContainsAny returns true if the set contains any of the values.
func (s *Set) ContainsAny(vals ...interface{}) bool {
	for _, v := range vals {
		if _, ok := s.m[v]; ok {
			return true
		}
	}
	return false
}

// Diff returns new Set about the values which other sets don't contain.
func (s Set) Diff(other Set) Set {
	res := NewSetWithSize(s.Size())

	for val := range s.m {
		if _, ok := other.m[val]; !ok {
			res.m[val] = struct{}{}
		}
	}
	return res
}

// DiffSlice is similar to Diff, but takes a slice as parameter.
func (s Set) DiffSlice(other []interface{}) Set {
	if len(s.m) > len(other) {
		tmp := NewSetWithSize(len(other))
		dup := 0
		for _, val := range other {
			if _, ok := s.m[val]; ok {
				dup++
			}
			tmp.m[val] = struct{}{}
		}
		res := NewSetWithSize(max(s.Size()-dup, 0))
		for val := range s.m {
			if _, ok := tmp.m[val]; !ok {
				res.m[val] = struct{}{}
			}
		}
		return res
	} else {
		res := NewSetWithSize(s.Size())
		for val := range s.m {
			res.m[val] = struct{}{}
		}
		for _, val := range other {
			if _, ok := res.m[val]; ok {
				delete(res.m, val)
			}
		}
		return res
	}
}

// FilterInclude returns a new slice which contains values that present in
// the provided slice and also present in the Set.
func (s Set) FilterInclude(slice []interface{}) []interface{} {
	res := make([]interface{}, 0, min(s.Size(), len(slice)))
	for _, val := range slice {
		if _, ok := s.m[val]; ok {
			res = append(res, val)
		}
	}
	return res
}

// FilterExclude returns a new slice which contains values that present in
// the provided slice but don't present in the Set.
func (s Set) FilterExclude(slice []interface{}) []interface{} {
	res := make([]interface{}, 0, len(slice))
	for _, val := range slice {
		if _, ok := s.m[val]; !ok {
			res = append(res, val)
		}
	}
	return res
}

// Intersect returns new Set about values which other set also contains.
func (s Set) Intersect(other Set) Set {
	res := NewSetWithSize(min(s.Size(), other.Size()))

	// loop over the smaller set
	if len(s.m) <= len(other.m) {
		for val := range s.m {
			if _, ok := other.m[val]; ok {
				res.m[val] = struct{}{}
			}
		}
	} else {
		for val := range other.m {
			if _, ok := s.m[val]; ok {
				res.m[val] = struct{}{}
			}
		}
	}
	return res
}

// IntersectSlice is similar to Intersect, but takes a slice as parameter.
func (s Set) IntersectSlice(other []interface{}) Set {
	res := NewSetWithSize(min(s.Size(), len(other)))

	for _, val := range other {
		if _, ok := s.m[val]; ok {
			res.m[val] = struct{}{}
		}
	}
	return res
}

// Union returns new Set about values either in the set or the other set.
func (s Set) Union(other Set) Set {
	res := NewSetWithSize(s.Size() + other.Size())

	for val := range s.m {
		res.m[val] = struct{}{}
	}
	for val := range other.m {
		res.m[val] = struct{}{}
	}
	return res
}

// UnionSlice is similar to Union, but takes a slice as parameter.
func (s Set) UnionSlice(other []interface{}) Set {
	res := NewSetWithSize(s.Size() + len(other))

	for val := range s.m {
		res.m[val] = struct{}{}
	}
	for _, val := range other {
		res.m[val] = struct{}{}
	}
	return res
}

// Slice converts set into interface{} slice.
func (s Set) Slice() []interface{} {
	res := make([]interface{}, 0, len(s.m))
	for val := range s.m {
		res = append(res, val)
	}
	return res
}

// SliceTo copy the set elements to the given dst slice.
//
// The param dst must be a pointer to either an interface slice, or a
// slice of the concrete element type, else it panics.
func (s Set) SliceTo(dst interface{}) {
	dstTyp := reflect.TypeOf(dst)
	if dstTyp == nil || dstTyp.Kind() != reflect.Ptr || dstTyp.Elem().Kind() != reflect.Slice {
		panic(fmt.Sprintf("invalid destination type %T", dst))
	}
	dstPtr := reflect.ValueOf(dst)
	dstElem := dstPtr.Elem()
	dstVal := dstElem
	if !dstElem.IsValid() {
		panic(fmt.Sprintf("invalid destination value %v", dst))
	}
	for val := range s.m {
		dstVal = reflect.Append(dstVal, reflect.ValueOf(val))
	}
	dstElem.Set(dstVal)
}

// Map converts set into map[interface{}]bool.
func (s Set) Map() map[interface{}]bool {
	res := make(map[interface{}]bool, len(s.m))
	for val := range s.m {
		res[val] = true
	}
	return res
}

// MapTo copy the set elements to the given map as keys.
//
// The param dst must be a pointer to either map[interface{}]bool, or a
// map using the concrete element type as key, else it panics.
func (s Set) MapTo(dst interface{}) {
	dstTyp := reflect.TypeOf(dst)
	if dstTyp == nil || dstTyp.Kind() != reflect.Ptr || dstTyp.Elem().Kind() != reflect.Map {
		panic(fmt.Sprintf("invalid destination type %T", dst))
	}
	dstPtr := reflect.ValueOf(dst)
	dstElem := dstPtr.Elem()
	dstVal := dstElem
	if !dstElem.IsValid() {
		panic(fmt.Sprintf("invalid destination value %v", dst))
	}
	trueVal := reflect.ValueOf(true)
	for val := range s.m {
		dstVal.SetMapIndex(reflect.ValueOf(val), trueVal)
	}
	dstElem.Set(dstVal)
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
