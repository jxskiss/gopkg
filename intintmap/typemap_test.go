package intintmap

import (
	"reflect"
	"testing"
	"unsafe"
)

func TestTypeMap(t *testing.T) {
	m := NewTypeMap()
	values1 := []interface{}{
		TestType1{},
		TestType2{},
		TestType3{},
		TestType4{},
		TestType5{},
		TestType6{},
	}
	values2 := []interface{}{
		TestType7{},
		TestType8{},
		TestType9{},
		TestType10{},
		TestType11{},
		TestType12{},
	}

	for _, val := range values1[:3] {
		//t.Logf("SetByType: %T", val)
		m.SetByType(reflect.TypeOf(val), 1)
	}
	for _, val := range values1[3:] {
		//t.Logf("SetByUintptr: %T", val)
		typeptr := (*(*[2]uintptr)(unsafe.Pointer(&val)))[0]
		m.SetByUintptr(typeptr, 1)
	}

	for _, val := range values1 {
		//t.Logf("GetByType: %T", val)
		got := m.GetByType(reflect.TypeOf(val))
		if got != 1 {
			t.Errorf("expected 1 as value, got %v", got)
		}
	}
	for _, val := range values2 {
		//t.Logf("GetByType: %T", val)
		got := m.GetByType(reflect.TypeOf(val))
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	}
}

type TestType1 struct{ A int }
type TestType2 struct{ B int32 }
type TestType3 struct{ C int64 }
type TestType4 struct{ D int8 }
type TestType5 struct{ E int }
type TestType6 struct{ F int }
type TestType7 struct{ G string }
type TestType8 struct{ H []byte }
type TestType9 struct{ I string }
type TestType10 struct{ J uint }
type TestType11 struct{ K uint }
type TestType12 struct{ L uint }
