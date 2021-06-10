// Code generated by go generate at 2021-05-30T01:20:54+08:00; DO NOT EDIT.

package set

import "encoding/json"

// Int32 is int32 set collection.
// The zero value of Int32 is an empty instance ready to use. A zero Int32
// value shall not be copied, or it may result incorrect behavior.
type Int32 struct {
	m map[int32]struct{}
}

// NewInt32 creates Int32 instance.
func NewInt32(vals ...int32) Int32 {
	size := max(len(vals), minSize)
	set := Int32{
		m: make(map[int32]struct{}, size),
	}
	set.Add(vals...)
	return set
}

// NewInt32WithSize creates Int32 instance with given initial size.
func NewInt32WithSize(size int) Int32 {
	set := Int32{
		m: make(map[int32]struct{}, size),
	}
	return set
}

// Size returns the size of set.
func (s Int32) Size() int { return len(s.m) }

// Add adds values into the set.
func (s *Int32) Add(vals ...int32) {
	if s.m == nil {
		size := max(len(vals), minSize)
		s.m = make(map[int32]struct{}, size)
	}
	for idx := range vals {
		s.m[vals[idx]] = struct{}{}
	}
}

// Del deletes values from the set.
func (s *Int32) Del(vals ...int32) {
	for idx := range vals {
		delete(s.m, vals[idx])
	}
}

// Pop pops an element from the set, in no particular order.
func (s *Int32) Pop() int32 {
	for val := range s.m {
		delete(s.m, val)
		return val
	}
	return 0
}

// Iterate iterates the set in no particular order and call the given function
// for each set element.
func (s Int32) Iterate(fn func(int32)) {
	for val := range s.m {
		fn(val)
	}
}

// Contains returns true if the set contains all the values.
func (s Int32) Contains(vals ...int32) bool {
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
func (s Int32) ContainsAny(vals ...int32) bool {
	for _, v := range vals {
		if _, ok := s.m[v]; ok {
			return true
		}
	}
	return false
}

// Diff returns new Int32 about the values which other set doesn't contain.
func (s Int32) Diff(other Int32) Int32 {
	res := NewInt32WithSize(s.Size())

	for val := range s.m {
		if _, ok := other.m[val]; !ok {
			res.m[val] = struct{}{}
		}
	}
	return res
}

// DiffSlice is similar to Diff, but takes a slice as parameter.
func (s Int32) DiffSlice(other []int32) Int32 {
	if len(s.m) > len(other) {
		tmp := NewInt32WithSize(len(other))
		dup := 0
		for _, val := range other {
			if _, ok := s.m[val]; ok {
				dup++
			}
			tmp.m[val] = struct{}{}
		}
		res := NewInt32WithSize(max(s.Size()-dup, 0))
		for val := range s.m {
			if _, ok := tmp.m[val]; !ok {
				res.m[val] = struct{}{}
			}
		}
		return res
	} else {
		res := NewInt32WithSize(s.Size())
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
// the provided slice and also present in the Int32 set.
func (s Int32) FilterInclude(slice []int32) []int32 {
	res := make([]int32, 0, min(s.Size(), len(slice)))
	for _, val := range slice {
		if _, ok := s.m[val]; ok {
			res = append(res, val)
		}
	}
	return res
}

// FilterExclude returns a new slice which contains values that present in
// the provided slice but don't present in the Int32 set.
func (s Int32) FilterExclude(slice []int32) []int32 {
	res := make([]int32, 0, len(slice))
	for _, val := range slice {
		if _, ok := s.m[val]; !ok {
			res = append(res, val)
		}
	}
	return res
}

// Intersect returns new Int32 about values which other set also contains.
func (s Int32) Intersect(other Int32) Int32 {
	res := NewInt32WithSize(min(s.Size(), other.Size()))

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
func (s Int32) IntersectSlice(other []int32) Int32 {
	res := NewInt32WithSize(min(s.Size(), len(other)))

	for _, val := range other {
		if _, ok := s.m[val]; ok {
			res.m[val] = struct{}{}
		}
	}
	return res
}

// Union returns new Int32 about values either in the set or the other set.
func (s Int32) Union(other Int32) Int32 {
	res := NewInt32WithSize(s.Size() + other.Size())

	for val := range s.m {
		res.m[val] = struct{}{}
	}
	for val := range other.m {
		res.m[val] = struct{}{}
	}
	return res
}

// UnionSlice is similar to Union, but takes a slice as parameter.
func (s Int32) UnionSlice(other []int32) Int32 {
	res := NewInt32WithSize(s.Size() + len(other))

	for val := range s.m {
		res.m[val] = struct{}{}
	}
	for _, val := range other {
		res.m[val] = struct{}{}
	}
	return res
}

// Slice converts set into int32 slice.
func (s Int32) Slice() []int32 {
	res := make([]int32, 0, len(s.m))

	for val := range s.m {
		res = append(res, val)
	}
	return res
}

// Map converts set into map[int32]bool.
func (s Int32) Map() map[int32]bool {
	res := make(map[int32]bool, len(s.m))

	for val := range s.m {
		res[val] = true
	}
	return res
}

// MarshalJSON implements json.Marshaler interface, the set will be
// marshaled as an int32 array.
func (s Int32) MarshalJSON() ([]byte, error) {
	res := s.Slice()
	return json.Marshal(res)
}

// UnmarshalJSON implements json.Unmarshaler interface, it will unmarshal
// an int32 array to the set.
func (s *Int32) UnmarshalJSON(b []byte) error {
	vals := make([]int32, 0)
	err := json.Unmarshal(b, &vals)
	if err == nil {
		s.Add(vals...)
	}
	return err
}

// MarshalYAML implements yaml.Marshaler interface of the yaml package,
// the set will be marshaled as an int32 array.
func (s Int32) MarshalYAML() (interface{}, error) {
	res := s.Slice()
	return res, nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface of the yaml package,
// it will unmarshal an int32 array to the set.
func (s *Int32) UnmarshalYAML(unmarshal func(interface{}) error) error {
	vals := make([]int32, 0)
	err := unmarshal(&vals)
	if err == nil {
		s.Add(vals...)
	}
	return err
}
