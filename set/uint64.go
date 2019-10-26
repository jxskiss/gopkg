// Code generated by go generate at 2019-10-26T10:41:30+08:00; DO NOT EDIT.

package set

import "encoding/json"

// Uint64 is uint64 set collection.
type Uint64 struct {
	m map[uint64]struct{}
}

// NewUint64 creates Uint64 instance.
func NewUint64(vals ...uint64) *Uint64 {
	size := max(len(vals), minSize)
	set := &Uint64{
		m: make(map[uint64]struct{}, size),
	}
	set.Add(vals...)
	return set
}

func NewUint64Size(size int) *Uint64 {
	set := &Uint64{
		m: make(map[uint64]struct{}, size),
	}
	return set
}

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
func (s *Uint64) Diff(other *Uint64) *Uint64 {
	res := NewUint64Size(s.Size())

	for val := range s.m {
		if !other.Contains(val) {
			res.Add(val)
		}
	}
	return res
}

// Intersect returns new Uint64 about values which other set also contains.
func (s *Uint64) Intersect(other *Uint64) *Uint64 {
	res := NewUint64Size(min(s.Size(), other.Size()))

	// loop over the smaller set
	if len(s.m) <= len(other.m) {
		for val := range s.m {
			if other.Contains(val) {
				res.Add(val)
			}
		}
	} else {
		for val := range other.m {
			if s.Contains(val) {
				res.Add(val)
			}
		}
	}
	return res
}

// Union returns new Uint64 about values either in the set or the other set.
func (s *Uint64) Union(other *Uint64) *Uint64 {
	res := NewUint64Size(s.Size() + other.Size())

	for val := range s.m {
		res.Add(val)
	}
	for val := range other.m {
		res.Add(val)
	}
	return res
}

// Size returns the size of set.
func (s *Uint64) Size() int {
	return len(s.m)
}

// Slice converts set into uint64 slice.
func (s *Uint64) Slice() []uint64 {
	res := make([]uint64, 0, len(s.m))

	for val := range s.m {
		res = append(res, val)
	}
	return res
}

// Map converts set into map[uint64]bool.
func (s *Uint64) Map() map[uint64]bool {
	res := make(map[uint64]bool, len(s.m))

	for val := range s.m {
		res[val] = true
	}
	return res
}

// MarshalJSON implements json.Marshaler interface, the set will be
// marshaled as an uint64 array.
func (s *Uint64) MarshalJSON() ([]byte, error) {
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
