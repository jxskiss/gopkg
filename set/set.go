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

func (s *Set) Pop() interface{} {
	for val := range s.m {
		delete(s.m, val)
		return val
	}
	return nil
}

func (s *Set) Each(fn func(interface{})) {
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

// Slice copy the keys to the given dst, which should be a slice of interface
// or an pointer to a slice of the key type.
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

	// interface slice
	if dstTyp.Elem().Elem().Kind() == reflect.Interface {
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

	// concrete slice type
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

	// interface map
	if dstTyp.Elem().Key().Kind() == reflect.Interface {
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

	// concrete map type
	if !dstVal.IsValid() || dstVal.IsNil() {
		dstVal = reflect.MakeMapWithSize(dstTyp.Elem(), len(s.m))
	}
	trueVal := reflect.ValueOf(true)
	for val := range s.m {
		dstVal.SetMapIndex(reflect.ValueOf(val), trueVal)
	}
	if dstElem.IsValid() {
		dstElem.Set(dstVal)
	}
	return dstVal.Interface()
}
