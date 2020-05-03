package set

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

type T1 struct{}

type T2 struct{}

func TestInt64Keys(t *testing.T) {
	target := []int64{1}
	values := []interface{}{
		map[int64]string{1: "a"},
		map[int64]int64{1: 2},
		map[int64]int{1: 3},
		map[int64]bool{1: true},
		map[int64]struct{}{1: {}},
		map[int64]map[int64]bool{1: {}},
		map[int64][]int64{1: {}},
		map[int64][]string{1: {}},
		map[int64]*T1{1: nil},
		map[int64]*T2{1: {}},
		map[int64][]*T1{1: {}},
	}

	for _, v := range values {
		if !reflect.DeepEqual(Int64Keys(v), target) {
			t.Errorf("failed: v = %T (%v)", v, v)
		}
	}
}

func TestStringKeys(t *testing.T) {
	target := []string{"a"}
	values := []interface{}{
		map[string]string{"a": "b"},
		map[string]int64{"a": 2},
		map[string]int{"a": 3},
		map[string]bool{"a": true},
		map[string]struct{}{"a": {}},
		map[string]map[int64]bool{"a": {}},
		map[string][]int64{"a": {}},
		map[string][]string{"a": {}},
		map[string]*T1{"a": nil},
		map[string]*T2{"a": {}},
		map[string][]*T1{"a": {}},
	}

	for _, v := range values {
		if !reflect.DeepEqual(StringKeys(v), target) {
			t.Errorf("failed: v = %T (%v)", v, v)
		}
	}
}

func TestSet_Add(t *testing.T) {
	t1 := map[interface{}]struct{}{
		1: {},
		2: {},
		3: {},
	}
	set1 := NewSet(1, 2, 3)
	set2 := NewSet([]int{1, 2, 3})
	if !reflect.DeepEqual(t1, set1.m) {
		t.Errorf("failed: set1")
	}
	if !reflect.DeepEqual(t1, set2.m) {
		t.Errorf("failec: set2")
	}
	set3 := NewSet()
	set3.Add(1, 2, 3)
	if !reflect.DeepEqual(t1, set3.m) {
		t.Errorf("failed set3")
	}
	set4 := NewSet()
	set4.Add([]int{1, 2, 3})
	if !reflect.DeepEqual(t1, set4.m) {
		t.Errorf("failed set4")
	}
}

func TestSet_Zero_Add(t *testing.T) {
	var set1 Set
	set1.Add(1, 2, 3)
	if set1.Size() != 3 {
		t.Errorf("failed add to zero set")
	}
}

func TestSet_Slice(t *testing.T) {
	set1 := NewSet(1, 2, 3)
	set1.Add([]int{4, 5, 6})
	target := []interface{}{1, 2, 3, 4, 5, 6}

	vals := set1.Slice()
	sort.Slice(vals, func(i, j int) bool {
		return vals[i].(int) < vals[j].(int)
	})

	if !reflect.DeepEqual(target, vals) {
		t.Errorf("failed Slice")
	}
}

func TestSet_SliceTo_Interface(t *testing.T) {
	set1 := NewSet(1, 2, 3)
	set1.Add([]int{4, 5, 6})
	target := []interface{}{1, 2, 3, 4, 5, 6}

	vals1 := make([]interface{}, 0)
	set1.SliceTo(&vals1)
	sort.Slice(vals1, func(i, j int) bool {
		return vals1[i].(int) < vals1[j].(int)
	})

	if !reflect.DeepEqual(target, vals1) {
		t.Errorf("failed SliceTo_Interface vals1")
	}
}

func TestSet_SliceTo_Int(t *testing.T) {
	set1 := NewSet(1, 2, 3)
	set1.Add([]int{4, 5, 6})
	target := []int{1, 2, 3, 4, 5, 6}

	vals1 := make([]int, 0)
	set1.SliceTo(&vals1)
	sort.Slice(vals1, func(i, j int) bool {
		return vals1[i] < vals1[j]
	})

	if !reflect.DeepEqual(target, vals1) {
		t.Errorf("failed SliceTo_Int vals1")
	}

	var vals3 []int
	set1.SliceTo(&vals3)
	sort.Slice(vals3, func(i, j int) bool {
		return vals3[i] < vals3[j]
	})

	if !reflect.DeepEqual(target, vals3) {
		t.Errorf("failed SliceTo_Int vals3")
	}
}

func TestSet_SliceTo_InvalidDst(t *testing.T) {
	set1 := NewSet(1, 2, 3)
	set1.Add([]int{4, 5, 6})

	// nil interface{}, should panic
	var vals1, vals2 interface{}
	err1 := shouldPanic(func() {
		set1.SliceTo(vals1)
	})
	if err1 == nil {
		t.Errorf("failed SliceTo_InvalidDst vals1")
	}
	err2 := shouldPanic(func() {
		set1.SliceTo(&vals2)
	})
	if err2 == nil {
		t.Errorf("failed SliceTo_InvalidDst vals2")
	}

	// （*[]interface{})(nil), should panic
	var vals3 *[]interface{}
	err3 := shouldPanic(func() {
		set1.SliceTo(vals3)
	})
	if err3 == nil {
		t.Errorf("failed SliceTo_InvalidDst vals3")
	}

	// (*[]int)(nil), should panic
	var vals4 *[]int
	err4 := shouldPanic(func() {
		set1.SliceTo(vals4)
	})
	if err4 == nil {
		t.Errorf("failed SliceTo_InvalidDst vals4")
	}
}

func TestSet_Map(t *testing.T) {
	set1 := NewSet(1, 2, 3)
	set1.Add([]int{4, 5, 6})
	target := map[interface{}]bool{
		1: true,
		2: true,
		3: true,
		4: true,
		5: true,
		6: true,
	}

	vals := set1.Map()
	if !reflect.DeepEqual(target, vals) {
		t.Errorf("failed Map")
	}
}

func TestSet_MapTo_Interface(t *testing.T) {
	set1 := NewSet(1, 2, 3)
	set1.Add([]int{4, 5, 6})
	target := map[interface{}]bool{
		1: true,
		2: true,
		3: true,
		4: true,
		5: true,
		6: true,
	}

	vals1 := make(map[interface{}]bool)
	set1.MapTo(&vals1)
	if !reflect.DeepEqual(target, vals1) {
		t.Errorf("failed MapTo_Interface vals1")
	}
}

func TestSet_MapTo_Int(t *testing.T) {
	set1 := NewSet(1, 2, 3)
	set1.Add(4, 5, 6)
	target := map[int]bool{
		1: true,
		2: true,
		3: true,
		4: true,
		5: true,
		6: true,
	}

	vals1 := make(map[int]bool)
	set1.MapTo(&vals1)
	if !reflect.DeepEqual(target, vals1) {
		t.Errorf("failed MapTo_Int vals1")
	}
}

func TestSet_MapTo_InvalidDst(t *testing.T) {
	set1 := NewSet(1, 2, 3)
	set1.Add(4, 5, 6)

	// nil interface{}, should panic
	var vals1, vals2 interface{}
	err1 := shouldPanic(func() {
		set1.MapTo(vals1)
	})
	if err1 == nil {
		t.Errorf("failed MapTo_InavlidDst vals1")
	}
	err2 := shouldPanic(func() {
		set1.MapTo(&vals2)
	})
	if err2 == nil {
		t.Errorf("failed MapTo_InvalidDst vals2")
	}

	// (*map[interface{}]bool)(nil), should panic
	var vals3 map[interface{}]bool
	err3 := shouldPanic(func() {
		set1.MapTo(&vals3)
	})
	if err3 == nil {
		t.Errorf("failed MapTo_InvalidDst vals3")
	}

	// (*map[int]bool)(nil), should panic
	var vals4 map[int]bool
	err4 := shouldPanic(func() {
		set1.MapTo(&vals4)
	})
	if err4 == nil {
		t.Errorf("failed MapTo_InvalidDst vals4")
	}
}

func TestSet_Chaining(t *testing.T) {
	set := NewSet(1, 2, 3, 4).
		Diff(NewSet(1, 2)).
		Union(NewSet(7, 8)).
		Intersect(NewSet(7, 8, 9, 0))
	if !reflect.DeepEqual(set.m, NewSet(7, 8).m) {
		t.Errorf("failed TestSet_Chaining")
	}
}

func shouldPanic(f func()) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()
	f()
	return
}
