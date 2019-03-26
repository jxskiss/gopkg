package set

import "encoding/json"

// Int64 is int64 set collection.
type Int64 struct {
	m map[int64]struct{}
}

// NewInt64 creates Int64 instance.
func NewInt64(vals ...int64) *Int64 {
	set := &Int64{
		m: make(map[int64]struct{}),
	}

	set.Add(vals...)
	return set
}

// Add adds values into the set.
func (s *Int64) Add(vals ...int64) {
	for idx := range vals {
		s.m[vals[idx]] = struct{}{}
	}
}

// Del delete values from the set.
func (s *Int64) Del(vals ...int64) {
	for idx := range vals {
		delete(s.m, vals[idx])
	}
}

// Pop pop an element from the set, in no particular order.
func (s *Int64) Pop() int64 {
	for val := range s.m {
		delete(s.m, val)
		return val
	}
	return 0
}

// Each iterate the set in no particular order and call the given function
// for each set element.
func (s *Int64) Each(fn func(int64)) {
	for val := range s.m {
		fn(val)
	}
}

// Has return true if the set has the value.
func (s *Int64) Has(val int64) bool {
	_, ok := s.m[val]
	return ok
}

// Diff return new Int64 about the values which other set doesn't contain.
func (s *Int64) Diff(other *Int64) *Int64 {
	res := NewInt64()

	for val := range s.m {
		if !other.Has(val) {
			res.Add(val)
		}
	}
	return res
}

// Intersect return new Int64 about values which other set also contains.
func (s *Int64) Intersect(other *Int64) *Int64 {
	res := NewInt64()

	// loop over the smaller set
	if len(s.m) <= len(other.m) {
		for val := range s.m {
			if other.Has(val) {
				res.Add(val)
			}
		}
	} else {
		for val := range other.m {
			if s.Has(val) {
				res.Add(val)
			}
		}
	}
	return res
}

// Union return new Int64 about values either in the set or the other set.
func (s *Int64) Union(other *Int64) *Int64 {
	res := NewInt64()

	for val := range s.m {
		res.Add(val)
	}
	for val := range other.m {
		res.Add(val)
	}
	return res
}

// Size return the size of set.
func (s *Int64) Size() int {
	return len(s.m)
}

// Slice converts set into int64 slice.
func (s *Int64) Slice() []int64 {
	res := make([]int64, 0, len(s.m))

	for val := range s.m {
		res = append(res, val)
	}
	return res
}

// Map converts set into map[int64]bool.
func (s *Int64) Map() map[int64]bool {
	res := make(map[int64]bool, len(s.m))

	for val := range s.m {
		res[val] = true
	}
	return res
}

// MarshalJSON implements json.Marshaler interface, the set will be
// marshaled as an int64 array.
func (s *Int64) MarshalJSON() ([]byte, error) {
	res := s.Slice()
	return json.Marshal(res)
}

// UnmarshalJSON implements json.Unmarshaler interface, it will unmarshal
// an int64 array to the set.
func (s *Int64) UnmarshalJSON(b []byte) error {
	vals := make([]int64, 0)
	err := json.Unmarshal(b, &vals)
	if err == nil {
		s.Add(vals...)
	}
	return err
}
