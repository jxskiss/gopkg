package set

import "encoding/json"

// String is string set collection.
type String struct {
	m map[string]struct{}
}

// NewString creates String instance.
func NewString(vals ...string) *String {
	set := &String{
		m: make(map[string]struct{}),
	}

	set.Add(vals...)
	return set
}

// Add adds values into the set.
func (s *String) Add(vals ...string) {
	for idx := range vals {
		s.m[vals[idx]] = struct{}{}
	}
}

// Del delete values from the set.
func (s *String) Del(vals ...string) {
	for idx := range vals {
		delete(s.m, vals[idx])
	}
}

// Pop pop an element from the set, in no particular order.
func (s *String) Pop() string {
	for val := range s.m {
		delete(s.m, val)
		return val
	}
	return ""
}

// Iterate iterate the set in no particular order and call the given function
// for each set element.
func (s *String) Iterate(fn func(string)) {
	for val := range s.m {
		fn(val)
	}
}

// Has return true if the set has the value.
func (s *String) Has(val string) bool {
	_, ok := s.m[val]
	return ok
}

// Diff return new String about the values which other set doesn't contain.
func (s *String) Diff(other *String) *String {
	res := NewString()

	for val := range s.m {
		if !other.Has(val) {
			res.Add(val)
		}
	}
	return res
}

// Intersect return new String about values which other set also contains.
func (s *String) Intersect(other *String) *String {
	res := NewString()

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

// Union return new String about values either in the set or the other set.
func (s *String) Union(other *String) *String {
	res := NewString()

	for val := range s.m {
		res.Add(val)
	}
	for val := range other.m {
		res.Add(val)
	}
	return res
}

// Size return the size of set.
func (s *String) Size() int {
	return len(s.m)
}

// Slice converts set into string slice.
func (s *String) Slice() []string {
	res := make([]string, 0, len(s.m))

	for val := range s.m {
		res = append(res, val)
	}
	return res
}

// Map converts the set into map[string]bool.
func (s *String) Map() map[string]bool {
	res := make(map[string]bool, len(s.m))

	for val := range s.m {
		res[val] = true
	}
	return res
}

// MarshalJSON implements json.Marshaler interface, the set will be
// marshaled as an string array.
func (s *String) MarshalJSON() ([]byte, error) {
	res := s.Slice()
	return json.Marshal(res)
}

// UnmarshalJSON implements json.Unmarshaler interface, it will unmarshal
// an string array to the set.
func (s *String) UnmarshalJSON(b []byte) error {
	vals := make([]string, 0)
	err := json.Unmarshal(b, &vals)
	if err == nil {
		s.Add(vals...)
	}
	return err
}
