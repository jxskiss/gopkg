package set

import "encoding/json"

// Generic is a generic set collection.
// The zero value of Generic is an empty instance ready to use.
// A zero Generic set value shall not be copied, or it may result
// incorrect behavior.
type Generic[T comparable] struct {
	m map[T]struct{}
}

// New creates a set instance and add the given values into the set.
func New[T comparable](vals ...T) Generic[T] {
	size := max(len(vals), minSize)
	set := Generic[T]{
		m: make(map[T]struct{}, size),
	}
	for _, v := range vals {
		set.m[v] = struct{}{}
	}
	return set
}

// NewWithSize creates a set instance with given initial size.
func NewWithSize[T comparable](size int) Generic[T] {
	set := Generic[T]{
		m: make(map[T]struct{}, size),
	}
	return set
}

// Size returns the size of the set collection.
func (s Generic[T]) Size() int { return len(s.m) }

// Add adds the given values into the set.
func (s *Generic[T]) Add(vals ...T) {
	if s.m == nil {
		size := max(len(vals), minSize)
		s.m = make(map[T]struct{}, size)
	}
	for _, v := range vals {
		s.m[v] = struct{}{}
	}
}

// Delete deletes values from the set.
func (s *Generic[T]) Delete(vals ...T) {
	for _, v := range vals {
		delete(s.m, v)
	}
}

// Iterate iterates the set in no particular order and calls the given
// function for each set element.
func (s Generic[T]) Iterate(fn func(T)) {
	for val := range s.m {
		fn(val)
	}
}

// Contains returns true if the set contains all the values.
func (s Generic[T]) Contains(vals ...T) bool {
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
func (s Generic[T]) ContainsAny(vals ...T) bool {
	for _, v := range vals {
		if _, ok := s.m[v]; ok {
			return true
		}
	}
	return false
}

// Diff returns a new set about the values which other set doesn't contain.
func (s Generic[T]) Diff(other Generic[T]) Generic[T] {
	res := NewWithSize[T](s.Size())

	for val := range s.m {
		if _, ok := other.m[val]; !ok {
			res.m[val] = struct{}{}
		}
	}
	return res
}

// DiffSlice is similar to Diff, but takes a slice as parameter.
func (s Generic[T]) DiffSlice(other []T) Generic[T] {
	otherLen := len(other)
	if len(s.m) > otherLen {
		tmp := NewWithSize[T](otherLen)
		dup := 0
		for i := 0; i < otherLen; i++ {
			val := other[i]
			if _, ok := s.m[val]; ok {
				dup++
			}
			tmp.m[val] = struct{}{}
		}
		res := NewWithSize[T](max(s.Size()-dup, 0))
		for val := range s.m {
			if _, ok := tmp.m[val]; !ok {
				res.m[val] = struct{}{}
			}
		}
		return res
	} else {
		res := NewWithSize[T](s.Size())
		for val := range s.m {
			res.m[val] = struct{}{}
		}
		for i := 0; i < otherLen; i++ {
			val := other[i]
			delete(res.m, val)
		}
		return res
	}
}

// FilterContains returns a new slice which contains values that present
// in the provided slice and also present in the set.
func (s Generic[T]) FilterContains(slice []T) []T {
	res := make([]T, 0, min(s.Size(), len(slice)))
	for _, val := range slice {
		if _, ok := s.m[val]; ok {
			res = append(res, val)
		}
	}
	return res
}

// FilterNotContains returns a new slice which contains values that present
// in the provided slice but don't present in the set.
func (s Generic[T]) FilterNotContains(slice []T) []T {
	res := make([]T, 0, len(slice))
	for _, val := range slice {
		if _, ok := s.m[val]; !ok {
			res = append(res, val)
		}
	}
	return res
}

// Intersect returns a new set about values which other set also contains.
func (s Generic[T]) Intersect(other Generic[T]) Generic[T] {
	res := NewWithSize[T](min(s.Size(), other.Size()))

	// loop over the smaller set for better performance
	small, big := s, other
	if s.Size() > other.Size() {
		small, big = other, s
	}
	for val := range small.m {
		if _, ok := big.m[val]; ok {
			res.m[val] = struct{}{}
		}
	}
	return res
}

// IntersectSlice is similar to Intersect, but takes a slice as parameter.
func (s Generic[T]) IntersectSlice(other []T) Generic[T] {
	res := NewWithSize[T](min(s.Size(), len(other)))
	for _, val := range other {
		if _, ok := s.m[val]; ok {
			res.m[val] = struct{}{}
		}
	}
	return res
}

// Union returns a new set about values either in the set or the other set.
func (s Generic[T]) Union(other Generic[T]) Generic[T] {
	res := NewWithSize[T](s.Size() + other.Size())
	for val := range s.m {
		res.m[val] = struct{}{}
	}
	for val := range other.m {
		res.m[val] = struct{}{}
	}
	return res
}

// UnionSlice is similar to Union, but takes a slice as parameter.
func (s Generic[T]) UnionSlice(other []T) Generic[T] {
	res := NewWithSize[T](s.Size() + len(other))
	for val := range s.m {
		res.m[val] = struct{}{}
	}
	for _, val := range other {
		res.m[val] = struct{}{}
	}
	return res
}

// Slice converts the set into a slice of type []T.
func (s Generic[T]) Slice() []T {
	res := make([]T, 0, len(s.m))
	for val := range s.m {
		res = append(res, val)
	}
	return res
}

// Map converts the set into a map of type map[T]bool.
func (s Generic[T]) Map() map[T]bool {
	res := make(map[T]bool, len(s.m))
	for val := range s.m {
		res[val] = true
	}
	return res
}

// MarshalJSON implements json.Marshaler interface, the set will be
// marshaled as a slice []T.
func (s Generic[T]) MarshalJSON() ([]byte, error) {
	res := s.Slice()
	return json.Marshal(res)
}

// UnmarshalJSON implements json.Unmarshaler interface, it will unmarshal
// a slice []T to the set.
func (s *Generic[T]) UnmarshalJSON(b []byte) error {
	vals := make([]T, 0)
	err := json.Unmarshal(b, &vals)
	if err == nil {
		s.Add(vals...)
	}
	return err
}

// MarshalYAML implements yaml.Marshaler interface of the yaml package,
// the set will be marshaled as a slice []T.
func (s Generic[T]) MarshalYAML() (any, error) {
	res := s.Slice()
	return res, nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface of the yaml package,
// it will unmarshal a slice []T to the set.
func (s *Generic[T]) UnmarshalYAML(unmarshal func(any) error) error {
	vals := make([]T, 0)
	err := unmarshal(&vals)
	if err == nil {
		s.Add(vals...)
	}
	return err
}
