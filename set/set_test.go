package set

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
)

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

func TestSet_DiffSlice(t *testing.T) {
	set1 := NewSet(1, 2, 3, 4)
	other := []int{2, 4, 5}
	got := set1.DiffSlice(other)
	if got.Size() != 2 {
		t.Errorf("unexpected set size after diff slice")
	}
	if !got.Contains(1, 3) {
		t.Errorf("unexpected set elements after diff slice")
	}
}

func TestSet_FilterInclude(t *testing.T) {
	set := NewSet(1, 2, 4, 6)
	slice := []int{2, 3, 4, 5}
	got := set.FilterContains(slice).([]int)
	want := []int{2, 4}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("failed filter include")
	}
}

func TestSet_FilterExclude(t *testing.T) {
	set := NewSet(1, 2, 4, 6)
	slice := []int{2, 3, 4, 5}
	got := set.FilterNotContains(slice).([]int)
	want := []int{3, 5}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("failed filter exclude")
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
