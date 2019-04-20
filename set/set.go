package set

import (
	"fmt"
	"reflect"
)

// Set is set collection of general type.
type Set struct {
	m map[interface{}]struct{}
}

// NewSet creates Set instance.
func NewSet(vals ...interface{}) *Set {
	set := &Set{
		m: make(map[interface{}]struct{}),
	}

	set.Add(vals...)
	return set
}

// Add adds values into set.
func (s *Set) Add(vals ...interface{}) {
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

// Del delete values from the set.
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

// Pop pop an element from the set, in no particular order.
func (s *Set) Pop() interface{} {
	for val := range s.m {
		delete(s.m, val)
		return val
	}
	return nil
}

// Iterate iterate the set in no particular order and call the given function
// for each set element.
func (s *Set) Iterate(fn func(interface{})) {
	for val := range s.m {
		fn(val)
	}
}

// Has return true if the set has the value.
func (s *Set) Has(val interface{}) bool {
	_, ok := s.m[val]
	return ok
}

// Diff return new Set about the values which other set doesn't contain.
func (s *Set) Diff(other *Set) *Set {
	res := NewSet()

	for val := range s.m {
		if !other.Has(val) {
			res.Add(val)
		}
	}
	return res
}

// Intersect return new Set about values which other set also contains.
func (s *Set) Intersect(other *Set) *Set {
	res := NewSet()

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

// Union return new Set about values either in the set or the other set.
func (s *Set) Union(other *Set) *Set {
	res := NewSet()

	for val := range s.m {
		res.Add(val)
	}
	for val := range other.m {
		res.Add(val)
	}
	return res
}

// Size return the size of the set.
func (s *Set) Size() int {
	return len(s.m)
}

// Slice copy the set elements to the given dst slice.
//
// The param dst should be either a pointer to an interface slice, or a
// pointer to a slice of the concrete element type, else it panics.
//
// The return value has same type with the slice that dst points to.
// Generally if dst is not nil, the return value should be ignored.
//
// If dst is a nil interface, the return value will be of type []interface{}
// which holds the set elements. And if dst is a pointer which point to a
// nil interface slice, the pointer will be pointed to the return value,
// which is a new malloced []interface{} holds the set elements.
//
// If dst is a pointer points to a concrete slice type, then the set elements
// will be appended to the slice, the return value is the same slice.
// In case of dst is a nil pointer, the return value will be a new malloced
// slice of the dst element type.
//
func (s *Set) Slice(dst interface{}) interface{} {
	if dst == nil {
		dstVal := make([]interface{}, 0, len(s.m))
		for val := range s.m {
			dstVal = append(dstVal, val)
		}
		return dstVal
	}

	dstTyp := reflect.TypeOf(dst)
	if dstTyp.Kind() != reflect.Ptr || dstTyp.Elem().Kind() != reflect.Slice {
		panic(fmt.Errorf("cannot convert set to destination: %T", dst))
	}

	dstPtr := reflect.ValueOf(dst)
	dstElem := dstPtr.Elem()
	dstVal := dstElem

	// interface slice or concrete slice type
	if !dstVal.IsValid() || dstVal.IsNil() {
		dstVal = reflect.MakeSlice(dstTyp.Elem(), 0, len(s.m))
	}
	for val := range s.m {
		dstVal = reflect.Append(dstVal, reflect.ValueOf(val))
	}
	if dstElem.IsValid() {
		dstElem.Set(dstVal)
	}
	return dstVal.Interface()
}

// Map copy the set elements to the given map as keys.
//
// The param dst should be either a pointer to map[interface{}]bool, or a
// pointer to a bool map using the concrete element type as key, else it panic.
//
// The return value has same type with the map that dst points to.
// Generally if dst is not nil, the return value should be ignored.
//
// If dst is a nil interface, the return value will be of type map[interface{}]bool,
// which holds the set elements as keys. And if dst is a pointer which point to
// a nil map[interface{}]bool, the pointer will be pointed to the return value,
// which is a new malloced map[interfaces{}]bool holds the set elements as keys.
//
// If dst is a pointer points to a map using the concrete element type as key,
// then the set elements will be populated into the map, the return value is
// the same map. In case of dst is a nil pointer, the return value will be a
// new malloced map of the same concrete type that dst points to.
//
func (s *Set) Map(dst interface{}) interface{} {
	if dst == nil {
		dstVal := make(map[interface{}]bool, len(s.m))
		for val := range s.m {
			dstVal[val] = true
		}
		return dstVal
	}

	dstTyp := reflect.TypeOf(dst)
	if dstTyp.Kind() != reflect.Ptr || dstTyp.Elem().Kind() != reflect.Map {
		panic(fmt.Errorf("cannot convert set to destination: %T", dst))
	}

	dstPtr := reflect.ValueOf(dst)
	dstElem := dstPtr.Elem()
	dstVal := dstElem

	// interface map or concrete map type
	trueVal := reflect.ValueOf(true)
	if !dstVal.IsValid() || dstVal.IsNil() {
		dstVal = reflect.MakeMapWithSize(dstTyp.Elem(), len(s.m))
	}
	for val := range s.m {
		dstVal.SetMapIndex(reflect.ValueOf(val), trueVal)
	}
	if dstElem.IsValid() {
		dstElem.Set(dstVal)
	}
	return dstVal.Interface()
}
