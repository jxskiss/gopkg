package set

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
func (s Set) Size() int { return len(s.m) }

// Add adds the given values into the set.
// If given only one param and which is a slice, the elements of the slice
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
func (s Set) Iterate(fn func(interface{})) {
	for val := range s.m {
		fn(val)
	}
}

// Contains returns true if the set contains all the values.
func (s Set) Contains(vals ...interface{}) bool {
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
func (s Set) ContainsAny(vals ...interface{}) bool {
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
// Param other must be a slice of []interface{} or slice of the concrete
// element type, else it panics.
func (s Set) DiffSlice(other interface{}) Set {
	otherTyp := reflect.TypeOf(other)
	if otherTyp == nil || otherTyp.Kind() != reflect.Slice {
		panic(fmt.Sprintf("invalid other type %T", other))
	}

	otherVal := reflect.ValueOf(other)
	otherLen := otherVal.Len()
	if len(s.m) > otherLen {
		tmp := NewSetWithSize(otherLen)
		dup := 0
		for i := 0; i < otherLen; i++ {
			val := otherVal.Index(i).Interface()
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
		for i := 0; i < otherLen; i++ {
			val := otherVal.Index(i).Interface()
			if _, ok := res.m[val]; ok {
				delete(res.m, val)
			}
		}
		return res
	}
}

// FilterInclude returns a new slice which contains values that present in
// the provided slice and also present in the Set.
// Param slice must be a slice of []interface{} or slice of the concrete
// element type, else it panics.
func (s Set) FilterInclude(slice interface{}) interface{} {
	sliceTyp := reflect.TypeOf(slice)
	if sliceTyp == nil || sliceTyp.Kind() != reflect.Slice {
		panic(fmt.Sprintf("invalid slice type %T", slice))
	}

	sliceVal := reflect.ValueOf(slice)
	sliceLen := sliceVal.Len()
	res := reflect.MakeSlice(sliceTyp, 0, min(s.Size(), sliceLen))
	for i := 0; i < sliceLen; i++ {
		val := sliceVal.Index(i)
		if _, ok := s.m[val.Interface()]; ok {
			res = reflect.Append(res, val)
		}
	}
	return res.Interface()
}

// FilterExclude returns a new slice which contains values that present in
// the provided slice but don't present in the Set.
// Param slice must be a slice of []interface{} or slice of the concrete
// element type, else it panics.
func (s Set) FilterExclude(slice interface{}) interface{} {
	sliceTyp := reflect.TypeOf(slice)
	if sliceTyp == nil || sliceTyp.Kind() != reflect.Slice {
		panic(fmt.Sprintf("invalid slice type %T", slice))
	}

	sliceVal := reflect.ValueOf(slice)
	sliceLen := sliceVal.Len()
	res := reflect.MakeSlice(sliceTyp, 0, sliceLen)
	for i := 0; i < sliceLen; i++ {
		val := sliceVal.Index(i)
		if _, ok := s.m[val.Interface()]; !ok {
			res = reflect.Append(res, val)
		}
	}
	return res.Interface()
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
// Param other must be a slice of []interface{} or slice of the concrete
// element type, else it panics.
func (s Set) IntersectSlice(other interface{}) Set {
	otherTyp := reflect.TypeOf(other)
	if otherTyp == nil || otherTyp.Kind() != reflect.Slice {
		panic(fmt.Sprintf("invalid other type %T", other))
	}

	otherVal := reflect.ValueOf(other)
	otherLen := otherVal.Len()
	res := NewSetWithSize(min(s.Size(), otherLen))
	for i := 0; i < otherLen; i++ {
		val := otherVal.Index(i).Interface()
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
// Param other must be a slice of []interface{} or slice of the concrete
// element type, else it panics.
func (s Set) UnionSlice(other interface{}) Set {
	otherTyp := reflect.TypeOf(other)
	if otherTyp == nil || otherTyp.Kind() != reflect.Slice {
		panic(fmt.Sprintf("invalid other type %T", other))
	}

	otherVal := reflect.ValueOf(other)
	otherLen := otherVal.Len()
	res := NewSetWithSize(s.Size() + otherLen)
	for val := range s.m {
		res.m[val] = struct{}{}
	}
	for i := 0; i < otherLen; i++ {
		val := otherVal.Index(i).Interface()
		res.m[val] = struct{}{}
	}
	return res
}

// Slice converts set into a []interface{} slice.
func (s Set) Slice() []interface{} {
	res := make([]interface{}, 0, len(s.m))
	for val := range s.m {
		res = append(res, val)
	}
	return res
}

// Map converts set into map[interface{}]bool.
func (s Set) Map() map[interface{}]bool {
	res := make(map[interface{}]bool, len(s.m))
	for val := range s.m {
		res[val] = true
	}
	return res
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
