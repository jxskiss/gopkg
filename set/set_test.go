package set

import (
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
		map[int64]*T2{1: &T2{}},
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
		map[string]*T2{"a": &T2{}},
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
		1: struct{}{},
		2: struct{}{},
		3: struct{}{},
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

func TestSet_Slice_Nil(t *testing.T) {
	set1 := NewSet(1, 2, 3)
	set1.Add([]int{4, 5, 6})
	target := []interface{}{1, 2, 3, 4, 5, 6}

	vals := set1.Slice(nil).([]interface{})
	sort.Slice(vals, func(i, j int) bool {
		return vals[i].(int) < vals[j].(int)
	})

	if !reflect.DeepEqual(target, vals) {
		t.Errorf("failed Slice_Nil")
	}
}

func TestSet_Slice_Interface(t *testing.T) {
	set1 := NewSet(1, 2, 3)
	set1.Add([]int{4, 5, 6})
	target := []interface{}{1, 2, 3, 4, 5, 6}

	vals1 := make([]interface{}, 0)
	vals2 := set1.Slice(&vals1).([]interface{})
	sort.Slice(vals1, func(i, j int) bool {
		return vals1[i].(int) < vals1[j].(int)
	})
	sort.Slice(vals2, func(i, j int) bool {
		return vals2[i].(int) < vals2[j].(int)
	})

	if !reflect.DeepEqual(target, vals1) {
		t.Errorf("failed Slice_Interface vals1")
	}
	if !reflect.DeepEqual(target, vals2) {
		t.Errorf("failed Slice_Interface vals2")
	}

	var vals3 = set1.Slice((*[]interface{})(nil)).([]interface{})
	sort.Slice(vals3, func(i, j int) bool {
		return vals3[i].(int) < vals3[j].(int)
	})
	if !reflect.DeepEqual(target, vals3) {
		t.Errorf("failed Slice_Interface vals3")
	}
}

func TestSet_Slice_Int(t *testing.T) {
	set1 := NewSet(1, 2, 3)
	set1.Add([]int{4, 5, 6})
	target := []int{1, 2, 3, 4, 5, 6}

	vals1 := make([]int, 0)
	vals2 := set1.Slice(&vals1).([]int)
	sort.Slice(vals1, func(i, j int) bool {
		return vals1[i] < vals1[j]
	})
	sort.Slice(vals2, func(i, j int) bool {
		return vals2[i] < vals2[j]
	})

	if !reflect.DeepEqual(target, vals1) {
		t.Errorf("failed Slice_Int vals1")
	}
	if !reflect.DeepEqual(target, vals2) {
		t.Errorf("failed Slice_Int_vals2")
	}

	var vals3 []int
	var vals4 = set1.Slice(&vals3).([]int)
	sort.Slice(vals3, func(i, j int) bool {
		return vals3[i] < vals3[j]
	})
	sort.Slice(vals4, func(i, j int) bool {
		return vals4[i] < vals4[j]
	})

	if !reflect.DeepEqual(target, vals3) {
		t.Errorf("failed Slice_Int vals3")
	}
	if !reflect.DeepEqual(target, vals4) {
		t.Errorf("failed Slice_Int vals4")
	}

	var vals6 = set1.Slice((*[]int)(nil)).([]int)
	sort.Slice(vals6, func(i, j int) bool {
		return vals6[i] < vals6[j]
	})
	if !reflect.DeepEqual(target, vals6) {
		t.Errorf("failed Slice_int, vals6")
	}
}

func TestSet_Map_Nil(t *testing.T) {
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

	vals := set1.Map(nil).(map[interface{}]bool)
	if !reflect.DeepEqual(target, vals) {
		t.Errorf("failed Map_Nil")
	}
}

func TestSet_Map_Interface(t *testing.T) {
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
	vals2 := set1.Map(&vals1).(map[interface{}]bool)
	if !reflect.DeepEqual(target, vals1) {
		t.Errorf("failed Map_Interface vals1")
	}
	if !reflect.DeepEqual(target, vals2) {
		t.Errorf("failed Map_Interface vals2")
	}

	var vals3 = set1.Map((*map[interface{}]bool)(nil)).(map[interface{}]bool)
	if !reflect.DeepEqual(target, vals3) {
		t.Errorf("failed Map_Interface vals3")
	}
}

func TestSet_Map_Int(t *testing.T) {
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
	vals2 := set1.Map(&vals1).(map[int]bool)
	if !reflect.DeepEqual(target, vals1) {
		t.Errorf("failed Map_Int vals1")
	}
	if !reflect.DeepEqual(target, vals2) {
		t.Errorf("failed Map_Int vals2")
	}

	var vals3 map[int]bool
	var vals4 = set1.Map(&vals3).(map[int]bool)
	if !reflect.DeepEqual(target, vals3) {
		t.Errorf("failed Map_Int vals3")
	}
	if !reflect.DeepEqual(target, vals4) {
		t.Errorf("failed Map_Int vals4")
	}

	var vals6 = set1.Map((*map[int]bool)(nil)).(map[int]bool)
	if !reflect.DeepEqual(target, vals6) {
		t.Errorf("failed Map_Int vals6")
	}
}
