package set

import (
	"reflect"
	"sort"
	"testing"
)

func TestGeneric_Add(t *testing.T) {
	t1 := map[int]struct{}{
		1: {},
		2: {},
		3: {},
	}
	set1 := New(1, 2, 3)
	if !reflect.DeepEqual(t1, set1.m) {
		t.Errorf("failed: set1")
	}
	set2 := New[int]()
	set2.Add(1, 2, 3)
	if !reflect.DeepEqual(t1, set2.m) {
		t.Errorf("failed set3")
	}
}

func TestGeneric_Zero_Add(t *testing.T) {
	var set1 Generic[int]
	set1.Add(1, 2, 3)
	if set1.Size() != 3 {
		t.Errorf("failed add to zero set")
	}
}

func TestGeneric_DiffSlice(t *testing.T) {
	set1 := New(1, 2, 3, 4)
	other := []int{2, 4, 5}
	got := set1.DiffSlice(other)
	if got.Size() != 2 {
		t.Errorf("unexpected set size after diff slice")
	}
	if !got.Contains(1, 3) {
		t.Errorf("unexpected set elements after diff slice")
	}
}

func TestGeneric_FilterInclude(t *testing.T) {
	set := New(1, 2, 4, 6)
	slice := []int{2, 3, 4, 5}
	got := set.FilterContains(slice)
	want := []int{2, 4}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("failed filter include")
	}
}

func TestGeneric_FilterExclude(t *testing.T) {
	set := New(1, 2, 4, 6)
	slice := []int{2, 3, 4, 5}
	got := set.FilterNotContains(slice)
	want := []int{3, 5}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("failed filter exclude")
	}
}

func TestGeneric_Slice(t *testing.T) {
	set1 := New(1, 2, 3)
	set1.Add([]int{4, 5, 6}...)
	target := []int{1, 2, 3, 4, 5, 6}

	vals := set1.Slice()
	sort.Slice(vals, func(i, j int) bool {
		return vals[i] < vals[j]
	})

	if !reflect.DeepEqual(target, vals) {
		t.Errorf("failed Slice")
	}
}

func TestGeneric_Map(t *testing.T) {
	set1 := New(1, 2, 3)
	set1.Add([]int{4, 5, 6}...)
	target := map[int]bool{
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

func TestGeneric_Chaining(t *testing.T) {
	set := New(1, 2, 3, 4).
		Diff(New(1, 2)).
		Union(New(7, 8)).
		Intersect(New(7, 8, 9, 0))
	if !reflect.DeepEqual(set.m, New(7, 8).m) {
		t.Errorf("failed TestGeneric_Chaining")
	}
}
