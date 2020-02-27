// Code generated by go generate at 2020-02-27T23:57:44+08:00; DO NOT EDIT.

package set

import "encoding/json"

// Uint64 is uint64 set collection.
// The zero value of Uint64 is an empty instance ready to use. A zero Uint64
// value shall not be copied, or it may result incorrect behavior.
type Uint64 struct {
	m map[uint64]struct{}
}

// NewUint64 creates Uint64 instance.
func NewUint64(vals ...uint64) Uint64 {
	size := max(len(vals), minSize)
	set := Uint64{
		m: make(map[uint64]struct{}, size),
	}
	set.Add(vals...)
	return set
}

// NewUint64WithSize creates Uint64 instance with given initial size.
func NewUint64WithSize(size int) Uint64 {
	set := Uint64{
		m: make(map[uint64]struct{}, size),
	}
	return set
}

// Size returns the size of set.
func (s *Uint64) Size() int { return len(s.m) }

// Add adds values into the set.
func (s *Uint64) Add(vals ...uint64) {
	if s.m == nil {
		size := max(len(vals), minSize)
		s.m = make(map[uint64]struct{}, size)
	}
	for idx := range vals {
		s.m[vals[idx]] = struct{}{}
	}
}

// Del deletes values from the set.
func (s *Uint64) Del(vals ...uint64) {
	for idx := range vals {
		delete(s.m, vals[idx])
	}
}

// Pop pops an element from the set, in no particular order.
func (s *Uint64) Pop() uint64 {
	for val := range s.m {
		delete(s.m, val)
		return val
	}
	return 0
}

// Iterate iterates the set in no particular order and call the given function
// for each set element.
func (s *Uint64) Iterate(fn func(uint64)) {
	for val := range s.m {
		fn(val)
	}
}

// Contains returns true if the set contains all the values.
func (s *Uint64) Contains(vals ...uint64) bool {
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
func (s *Uint64) ContainsAny(vals ...uint64) bool {
	for _, v := range vals {
		if _, ok := s.m[v]; ok {
			return true
		}
	}
	return false
}

// Diff returns new Uint64 about the values which other set doesn't contain.
func (s Uint64) Diff(other Uint64) Uint64 {
	res := NewUint64WithSize(s.Size())

	for val := range s.m {
		if _, ok := other.m[val]; !ok {
			res.m[val] = struct{}{}
		}
	}
	return res
}

// DiffSlice is similar to Diff, but takes a slice as parameter.
func (s Uint64) DiffSlice(other []uint64) Uint64 {
	if len(s.m) > len(other) {
		tmp := NewUint64WithSize(len(other))
		dup := 0
		for _, val := range other {
			if _, ok := s.m[val]; ok {
				dup++
			}
			tmp.m[val] = struct{}{}
		}
		res := NewUint64WithSize(max(s.Size()-dup, 0))
		for val := range s.m {
			if _, ok := tmp.m[val]; !ok {
				res.m[val] = struct{}{}
			}
		}
		return res
	} else {
		res := NewUint64WithSize(s.Size())
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
// the provided slice and also present in the Uint64 set.
func (s Uint64) FilterInclude(slice []uint64) []uint64 {
	res := make([]uint64, 0, min(s.Size(), len(slice)))
	for _, val := range slice {
		if _, ok := s.m[val]; ok {
			res = append(res, val)
		}
	}
	return res
}

// FilterExclude returns a new slice which contains values that present in
// the provided slice but don't present in the Uint64 set.
func (s Uint64) FilterExclude(slice []uint64) []uint64 {
	res := make([]uint64, 0, len(slice))
	for _, val := range slice {
		if _, ok := s.m[val]; !ok {
			res = append(res, val)
		}
	}
	return res
}

// Intersect returns new Uint64 about values which other set also contains.
func (s Uint64) Intersect(other Uint64) Uint64 {
	res := NewUint64WithSize(min(s.Size(), other.Size()))

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
func (s Uint64) IntersectSlice(other []uint64) Uint64 {
	res := NewUint64WithSize(min(s.Size(), len(other)))

	for _, val := range other {
		if _, ok := s.m[val]; ok {
			res.m[val] = struct{}{}
		}
	}
	return res
}

// Union returns new Uint64 about values either in the set or the other set.
func (s Uint64) Union(other Uint64) Uint64 {
	res := NewUint64WithSize(s.Size() + other.Size())

	for val := range s.m {
		res.m[val] = struct{}{}
	}
	for val := range other.m {
		res.m[val] = struct{}{}
	}
	return res
}

// UnionSlice is similar to Union, but takes a slice as parameter.
func (s Uint64) UnionSlice(other []uint64) Uint64 {
	res := NewUint64WithSize(s.Size() + len(other))

	for val := range s.m {
		res.m[val] = struct{}{}
	}
	for _, val := range other {
		res.m[val] = struct{}{}
	}
	return res
}

// Slice converts set into uint64 slice.
func (s Uint64) Slice() []uint64 {
	res := make([]uint64, 0, len(s.m))

	for val := range s.m {
		res = append(res, val)
	}
	return res
}

// Map converts set into map[uint64]bool.
func (s Uint64) Map() map[uint64]bool {
	res := make(map[uint64]bool, len(s.m))

	for val := range s.m {
		res[val] = true
	}
	return res
}

// MarshalJSON implements json.Marshaler interface, the set will be
// marshaled as an uint64 array.
func (s Uint64) MarshalJSON() ([]byte, error) {
	res := s.Slice()
	return json.Marshal(res)
}

// UnmarshalJSON implements json.Unmarshaler interface, it will unmarshal
// an uint64 array to the set.
func (s *Uint64) UnmarshalJSON(b []byte) error {
	vals := make([]uint64, 0)
	err := json.Unmarshal(b, &vals)
	if err == nil {
		s.Add(vals...)
	}
	return err
}
