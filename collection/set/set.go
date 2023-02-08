package set

import (
	"encoding/json"
	"fmt"
	"reflect"
)

const minSize = 8

// Set is a set collection of any type.
// The zero value of Set is an empty instance ready to use. A zero Set
// value shall not be copied, or it may result incorrect behavior.
type Set struct {
	m map[any]struct{}
}

// NewSet creates a Set instance and add the given values into the set.
// If given only one param which is a slice, the elements of the slice
// will be added into the set using reflection.
func NewSet(vals ...any) Set {
	size := max(len(vals), minSize)
	set := Set{
		m: make(map[any]struct{}, size),
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

// NewSetWithSize creates a Set instance with given initial size.
func NewSetWithSize(size int) Set {
	set := Set{
		m: make(map[any]struct{}, size),
	}
	return set
}

// Size returns the size of the set.
func (s Set) Size() int { return len(s.m) }

// Add adds the given values into the set.
// If given only one param which is a slice, the elements of the slice
// will be added into the set using reflection.
func (s *Set) Add(vals ...any) {
	if s.m == nil {
		size := max(len(vals), minSize)
		s.m = make(map[any]struct{}, size)
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
//
// Deprecated: Del has been renamed to Delete.
func (s *Set) Del(vals ...any) {
	s.Delete(vals...)
}

// Delete deletes values from the set.
func (s *Set) Delete(vals ...any) {
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

// Iterate iterates the set in no particular order and calls the given
// function for each set element.
func (s Set) Iterate(fn func(any)) {
	for val := range s.m {
		fn(val)
	}
}

// Contains returns true if the set contains all the values.
func (s Set) Contains(vals ...any) bool {
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
func (s Set) ContainsAny(vals ...any) bool {
	for _, v := range vals {
		if _, ok := s.m[v]; ok {
			return true
		}
	}
	return false
}

// Diff returns a new Set about the values which other sets don't contain.
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
// Param other must be a slice of []any or slice of the concrete
// element type, else it panics.
func (s Set) DiffSlice(other any) Set {
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
			delete(res.m, val)
		}
		return res
	}
}

// FilterInclude returns a new slice which contains values that present
// in the provided slice and also present in the Set.
// Param slice must be a slice of []any or slice of the concrete
// element type, else it panics.
//
// Deprecated: FilterInclude has been renamed to FilterContains.
func (s Set) FilterInclude(slice any) any {
	return s.FilterContains(slice)
}

// FilterContains returns a new slice which contains values that present
// in the provided slice and also present in the Set.
// Param slice must be a slice of []any or slice of the concrete
// element type, else it panics.
func (s Set) FilterContains(slice any) any {
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

// FilterExclude returns a new slice which contains values that present
// in the provided slice but don't present in the Set.
// Param slice must be a slice of []any or slice of the concrete
// element type, else it panics.
//
// Deprecated: FilterExclude has been renamed to FilterNotContains.
func (s Set) FilterExclude(slice any) any {
	return s.FilterNotContains(slice)
}

// FilterNotContains returns a new slice which contains values that present
// in the provided slice but don't present in the Set.
// Param slice must be a slice of []any or slice of the concrete
// element type, else it panics.
func (s Set) FilterNotContains(slice any) any {
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
// Param other must be a slice of []any or slice of the concrete
// element type, else it panics.
func (s Set) IntersectSlice(other any) Set {
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
// Param other must be a slice of []any or slice of the concrete
// element type, else it panics.
func (s Set) UnionSlice(other any) Set {
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

// Slice converts set into a slice of type []any.
func (s Set) Slice() []any {
	res := make([]any, 0, len(s.m))
	for val := range s.m {
		res = append(res, val)
	}
	return res
}

// Map converts set into a map of type map[any]bool.
func (s Set) Map() map[any]bool {
	res := make(map[any]bool, len(s.m))
	for val := range s.m {
		res[val] = true
	}
	return res
}

// MarshalJSON implements json.Marshaler interface, the set will be
// marshaled as a slice []any.
func (s Set) MarshalJSON() ([]byte, error) {
	res := s.Slice()
	return json.Marshal(res)
}

// UnmarshalJSON implements json.Unmarshaler interface, it will unmarshal
// a slice []any to the set.
func (s *Set) UnmarshalJSON(b []byte) error {
	vals := make([]any, 0)
	err := json.Unmarshal(b, &vals)
	if err == nil {
		s.Add(vals...)
	}
	return err
}

// MarshalYAML implements yaml.Marshaler interface of the yaml package,
// the set will be marshaled as a slice []any.
func (s Set) MarshalYAML() (any, error) {
	res := s.Slice()
	return res, nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface of the yaml package,
// it will unmarshal a slice []any to the set.
func (s *Set) UnmarshalYAML(unmarshal func(any) error) error {
	vals := make([]any, 0)
	err := unmarshal(&vals)
	if err == nil {
		s.Add(vals...)
	}
	return err
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
