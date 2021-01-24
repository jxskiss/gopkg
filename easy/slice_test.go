package easy

import (
	"github.com/jxskiss/gopkg/ptr"
	"github.com/stretchr/testify/assert"
	"testing"
)

type simple struct {
	A string
}

var insertSliceTests = []map[string]interface{}{
	{
		"func":  InsertSlice,
		"slice": []int64{1, 2, 3, 4},
		"elem":  int64(9),
		"index": 3,
		"want":  []int64{1, 2, 3, 9, 4},
	},
	{
		"func":  InsertSlice,
		"slice": []int64{1, 2, 3, 4},
		"elem":  int(9), // int type not match
		"index": 3,
		"want":  []int64{1, 2, 3, 9, 4},
	},
	{
		"func":  InsertSlice,
		"slice": []int{1, 2, 3, 4},
		"elem":  9,
		"index": 10,
		"want":  []int{1, 2, 3, 4, 9},
	},
	{
		"func":  InsertSlice,
		"slice": []string{"1", "2", "3", "4"},
		"elem":  "9",
		"index": 3,
		"want":  []string{"1", "2", "3", "9", "4"},
	},
	{
		"func":  InsertSlice,
		"slice": Strings{"1", "2", "3", "4"},
		"elem":  "9",
		"index": 10,
		"want":  Strings{"1", "2", "3", "4", "9"},
	},
	{
		"func":  InsertSlice,
		"slice": []simple{{"a"}, {"b"}, {"c"}, {"d"}},
		"elem":  simple{"z"},
		"index": 3,
		"want":  []simple{{"a"}, {"b"}, {"c"}, {"z"}, {"d"}},
	},
	{
		"func":  InsertSlice,
		"slice": []int64{1, 2, 3, 4},
		"elem":  int16(9), // int type not match
		"index": 10,       // exceeds slice length
		"want":  []int64{1, 2, 3, 4, 9},
	},
	{
		"func":  InsertSlice,
		"slice": nil,
		"elem":  9,
		"index": 3,
		"want":  "panic",
	},
	{
		"func":  InsertSlice,
		"slice": []int64{1, 2, 3, 4},
		"elem":  nil,
		"index": 3,
		"want":  "panic",
	},
	{
		"func":  InsertSlice,
		"slice": []*simple{},
		"elem":  nil,
		"index": 3,
		"want":  "panic",
	},
	{
		"func":  InsertSlice,
		"slice": []*simple{},
		"elem":  (*simple)(nil),
		"index": 3,
		"want":  []*simple{nil},
	},
}

func TestInsertSlice(t *testing.T) {
	for _, test := range insertSliceTests {
		f := test["func"].(func(slice interface{}, index int, elem interface{}) interface{})

		var got interface{}
		insert := func() {
			got = f(test["slice"], test["index"].(int), test["elem"])
		}
		if test["want"] == "panic" {
			assert.Panics(t, insert)
		} else {
			assert.NotPanics(t, insert)
			assert.Equal(t, test["want"], got)
		}
	}
}

type comptyp struct {
	I32   int32
	I32_p *int32

	I64   int64
	I64_p *int64

	Str   string
	Str_p *string

	Simple   simple
	Simple_p *simple
}

var complexTypeData = []*comptyp{
	{
		I32:      32,
		I32_p:    ptr.Int32(32),
		I64:      64,
		I64_p:    ptr.Int64(64),
		Str:      "str",
		Str_p:    ptr.String("str"),
		Simple:   simple{A: "a"},
		Simple_p: &simple{A: "a"},
	},
	{
		I32:      33,
		I32_p:    ptr.Int32(33),
		I64:      65,
		I64_p:    ptr.Int64(65),
		Str:      "str_2",
		Str_p:    ptr.String("str_2"),
		Simple:   simple{A: "b"},
		Simple_p: &simple{A: "b"},
	},
	{
		I32:      34,
		I32_p:    ptr.Int32(34),
		I64:      66,
		I64_p:    ptr.Int64(66),
		Str:      "str_3",
		Str_p:    ptr.String("str_3"),
		Simple:   simple{A: "c"},
		Simple_p: &simple{A: "c"},
	},
	{
		I32:      35,
		I32_p:    nil,
		I64:      67,
		I64_p:    nil,
		Str:      "str_4",
		Str_p:    nil,
		Simple:   simple{A: "d"},
		Simple_p: nil,
	},
}

func TestPluck(t *testing.T) {
	want := []string{"a", "b", "c"}
	slice1 := []simple{{"a"}, {"b"}, {"c"}}
	slice2 := []*simple{{"a"}, {"b"}, {"c"}}

	assert.Equal(t, want, Pluck(slice1, "A"))
	assert.Equal(t, want, Pluck(slice2, "A"))

	assert.Panics(t, func() { Pluck(&slice1, "A") })
	assert.Panics(t, func() { Pluck(&slice2, "A") })

	assert.Panics(t, func() { Pluck(nil, "A") })
	assert.Panics(t, func() { Pluck(slice1, "") })
}

func TestPluckStrings(t *testing.T) {
	want := Strings{"a", "b", "c"}
	slice1 := []simple{{"a"}, {"b"}, {"c"}}
	slice2 := []*simple{{"a"}, {"b"}, {"c"}}

	assert.Equal(t, want, PluckStrings(slice1, "A"))
	assert.Equal(t, want, PluckStrings(slice2, "A"))

	assert.Panics(t, func() { PluckStrings(&slice1, "A") })
	assert.Panics(t, func() { PluckStrings(&slice2, "A") })

	assert.Panics(t, func() { PluckStrings(nil, "A") })
	assert.Panics(t, func() { PluckStrings(slice1, "") })
}

func TestPluckInt64s(t *testing.T) {
	slice := complexTypeData

	got1 := PluckInt64s(slice, "I32")
	want1 := Int64s{32, 33, 34, 35}
	assert.Equal(t, want1, got1)

	got2 := PluckInt64s(slice, "I32_p")
	want2 := Int64s{32, 33, 34}
	assert.Equal(t, want2, got2)

	got3 := PluckInt64s(slice, "I64")
	want3 := Int64s{64, 65, 66, 67}
	assert.Equal(t, want3, got3)

	got4 := PluckInt64s(slice, "I64_p")
	want4 := Int64s{64, 65, 66}
	assert.Equal(t, want4, got4)

	assert.Panics(t, func() { PluckInt64s(nil, "I32") })
	assert.Panics(t, func() { PluckInt64s(slice, "") })
}

func TestPluck_StructField(t *testing.T) {
	slice := complexTypeData

	got1 := Pluck(slice, "Simple")
	want1 := []simple{{"a"}, {"b"}, {"c"}, {"d"}}
	assert.Equal(t, want1, got1)

	got2 := Pluck(slice, "Simple_p")
	assert.IsType(t, []*simple(nil), got2)
	assert.Len(t, got2, len(slice))
	assert.Nil(t, got2.([]*simple)[3])
}

var reverseSliceTests = []map[string]interface{}{
	{
		"slice": []uint64{1, 2, 3},
		"want":  []uint64{3, 2, 1},
	},
	{
		"slice": []int8{1, 2, 3},
		"want":  []int8{3, 2, 1},
	},
	{
		"slice": []string{"1", "2", "3"},
		"want":  []string{"3", "2", "1"},
	},
	{
		"slice": []simple{{"a"}, {"b"}, {"c"}},
		"want":  []simple{{"c"}, {"b"}, {"a"}},
	},
	{
		"slice": []int(nil),
		"want":  []int{},
	},
}

func TestReverseSlice(t *testing.T) {
	for _, test := range reverseSliceTests {
		got := ReverseSlice(test["slice"])
		assert.Equal(t, test["want"], got)
	}

	assert.Panics(t, func() { ReverseSlice(nil) })
}

var uniqueSliceTests = []map[string]interface{}{
	{
		"slice": []uint64{2, 2, 1, 3, 2, 3, 1, 3},
		"want":  []uint64{2, 1, 3},
	},
	{
		"slice": []int8{2, 2, 1, 3, 2, 3, 1, 3},
		"want":  []int8{2, 1, 3},
	},
	{
		"slice": []string{"2", "2", "1", "3", "2", "3", "1", "3"},
		"want":  []string{"2", "1", "3"},
	},
	{
		"slice": []simple{{"2"}, {"2"}, {"1"}, {"3"}, {"2"}, {"3"}, {"1"}, {"3"}},
		"want":  []simple{{"2"}, {"1"}, {"3"}},
	},
	{
		"slice": []int(nil),
		"want":  []int{},
	},
}

func TestUniqueSlice(t *testing.T) {
	for _, test := range uniqueSliceTests {
		got := UniqueSlice(test["slice"])
		assert.Equal(t, test["want"], got)
	}
}

func TestDiffInt64s(t *testing.T) {
	a := []int64{1, 2, 3, 4, 5}
	b := []int64{4, 5, 6, 7, 8}
	want1 := Int64s{1, 2, 3}
	assert.Equal(t, want1, DiffInt64s(a, b))

	want2 := Int64s{6, 7, 8}
	assert.Equal(t, want2, DiffInt64s(b, a))
}

func TestDiffStrings(t *testing.T) {
	a := []string{"1", "2", "3", "4", "5"}
	b := []string{"4", "5", "6", "7", "8"}
	want1 := Strings{"1", "2", "3"}
	assert.Equal(t, want1, DiffStrings(a, b))

	want2 := Strings{"6", "7", "8"}
	assert.Equal(t, want2, DiffStrings(b, a))
}

func TestToMap(t *testing.T) {
	a := &simple{"a"}
	b := &simple{"b"}
	c := &simple{"c"}
	slice := []*simple{a, b, c}
	want := map[string]*simple{"a": a, "b": b, "c": c}
	got := ToMap(slice, "A")
	assert.Equal(t, want, got)

	assert.Panics(t, func() { ToMap(nil, "A") })
	assert.Panics(t, func() { ToMap(slice, "") })
}

func TestToMap_Pointer(t *testing.T) {
	a := &comptyp{Str_p: ptr.String("a")}
	b := &comptyp{Str_p: ptr.String("b")}
	c := &comptyp{Str_p: ptr.String("c")}
	slice := []*comptyp{a, b, c}
	want := map[string]*comptyp{"a": a, "b": b, "c": c}
	got := ToMap(slice, "Str_p")
	assert.Equal(t, want, got)
}

func TestToSliceMap(t *testing.T) {
	a := &comptyp{I32: 1, I32_p: ptr.Int32(1)}
	b := &comptyp{I32: 1, I32_p: ptr.Int32(1)}
	c := &comptyp{I32: 2, I32_p: ptr.Int32(2)}

	slice1 := []comptyp{*a, *b, *c}
	want1 := map[int32][]comptyp{
		1: {*a, *b},
		2: {*c},
	}
	got1 := ToSliceMap(slice1, "I32").(map[int32][]comptyp)
	assert.Len(t, got1, len(want1))
	assert.ElementsMatch(t, MapKeys(got1), MapKeys(want1))
	assert.ElementsMatch(t, got1[1], want1[1])
	assert.ElementsMatch(t, got1[2], want1[2])

	want2 := want1
	got2 := ToSliceMap(slice1, "I32_p").(map[int32][]comptyp)
	assert.Len(t, got2, len(want1))
	assert.ElementsMatch(t, MapKeys(got2), MapKeys(want1))
	assert.ElementsMatch(t, got2[1], want2[1])
	assert.ElementsMatch(t, got2[2], want2[2])

	slice3 := []*comptyp{a, b, c}
	want3 := map[int32][]*comptyp{
		1: {a, b},
		2: {c},
	}
	got3 := ToSliceMap(slice3, "I32").(map[int32][]*comptyp)
	assert.Len(t, got3, len(want3))
	assert.ElementsMatch(t, MapKeys(got3), MapKeys(want3))
	assert.ElementsMatch(t, got3[1], want3[1])
	assert.ElementsMatch(t, got3[2], want3[2])

	want4 := want3
	got4 := ToSliceMap(slice3, "I32_p").(map[int32][]*comptyp)
	assert.Len(t, got4, len(want4))
	assert.ElementsMatch(t, MapKeys(got4), MapKeys(want4))
	assert.ElementsMatch(t, got4[1], want4[1])
	assert.ElementsMatch(t, got4[2], want4[2])

	// panics
	assert.Panics(t, func() { ToSliceMap(nil, "I32_p") })
	assert.Panics(t, func() { ToSliceMap(slice1, "") })
	assert.Panics(t, func() { ToSliceMap(a, "I32_p") })
}

func TestToMapMap(t *testing.T) {
	a := &comptyp{I32: 1, I32_p: ptr.Int32(1)}
	b := &comptyp{I32: 1, I32_p: ptr.Int32(2)}
	c := &comptyp{I32: 3, I32_p: ptr.Int32(3)}

	slice1 := []comptyp{*a, *b, *c}
	want1 := map[int32]map[int32]comptyp{
		1: {1: *a, 2: *b},
		3: {3: *c},
	}
	got1 := ToMapMap(slice1, "I32", "I32_p")
	assert.Equal(t, want1, got1)

	slice2 := []*comptyp{a, b, c}
	want2 := map[int32]map[int32]*comptyp{
		1: {1: a, 2: b},
		3: {3: c},
	}
	got2 := ToMapMap(slice2, "I32", "I32_p")
	assert.Equal(t, want2, got2)

	// panics
	assert.Panics(t, func() { ToMapMap(nil, "I32", "I32_p") })
	assert.Panics(t, func() { ToMapMap(slice1, "", "I32_p") })
	assert.Panics(t, func() { ToMapMap(slice1, "I32", "") })
	assert.Panics(t, func() { ToMapMap(a, "I32", "I32_p") })
}

func TestFindAndFilter(t *testing.T) {
	a := &comptyp{I32: 1, Str_p: ptr.String("a")}
	b := &comptyp{I64: 2, Str_p: ptr.String("b")}
	c := &comptyp{I64_p: ptr.Int64(3), Str_p: ptr.String("c")}
	slice := []*comptyp{a, b, c}

	f1 := func(i int) bool { return slice[i].Str_p == nil }
	got1 := Find(slice, f1)
	all1 := Filter(slice, f1)

	assert.Equal(t, nil, got1)
	assert.NotNil(t, all1)
	assert.Len(t, all1, 0)

	f3 := func(i int) bool { return ptr.DerefInt64(slice[i].I64_p) == 3 }
	got3 := Find(slice, f3)
	all3 := Filter(slice, f3)
	assert.Equal(t, c, got3)
	assert.Len(t, all3, 1)

	assert.Panics(t, func() { Find(slice, nil) })
	assert.Panics(t, func() { Filter(nil, f3) })
	assert.Panics(t, func() { Filter(slice, nil) })
}

func TestParseInt64s(t *testing.T) {
	strIntIDs := "123,,456,789, ,0,"
	want := Int64s{123, 456, 789}
	got, isMalformed := ParseInt64s(strIntIDs, ",", true)
	assert.True(t, isMalformed)
	assert.Equal(t, want, got)
}

func TestJoinInt64s(t *testing.T) {
	slice := []int64{1, 2, 3, 4, 5}
	want := "1,2,3,4,5"
	got := JoinInt64s(slice, ",")
	assert.Equal(t, want, got)
}

var splitBatchTests = []map[string]interface{}{
	{
		"total": 0,
		"batch": 10,
		"want":  []IJ(nil),
	},
	{
		"total": 72,
		"batch": -36,
		"want":  []IJ{{0, 72}},
	},
	{
		"total": 72,
		"batch": 0,
		"want":  []IJ{{0, 72}},
	},
	{
		"total": 72,
		"batch": 35,
		"want":  []IJ{{0, 35}, {35, 70}, {70, 72}},
	},
	{
		"total": 72,
		"batch": 24,
		"want":  []IJ{{0, 24}, {24, 48}, {48, 72}},
	},
}

func TestSplitBatch(t *testing.T) {
	for _, test := range splitBatchTests {
		got := SplitBatch(test["total"].(int), test["batch"].(int))
		assert.Equal(t, test["want"], got)
	}
}

func TestSplitSlice(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6, 7}
	want := [][]int{{1, 2, 3}, {4, 5, 6}, {7}}
	got := SplitSlice(slice, 3)
	assert.Equal(t, want, got)
}

func TestSumSlice(t *testing.T) {
	tests := []map[string]interface{}{
		{"slice": []int(nil), "sum": 0},
		{"slice": []int32{1, 2, 3}, "sum": 6},
		{"slice": []uint64{4, 5, 6}, "sum": 15},
	}
	for _, test := range tests {
		got := SumSlice(test["slice"])
		assert.Equal(t, test["sum"], int(got))
	}
}

func TestSumMapSlice(t *testing.T) {
	mSlice := map[string][]int{
		"a": {1, 2, 3},
		"b": {4, 5, 6},
	}
	got := SumMapSlice(mSlice)
	want := int64(21)
	assert.Equal(t, want, got)
}

func TestSumMapSliceLength(t *testing.T) {
	mSlice := map[string][]int{
		"a": {1, 2, 3},
		"b": {4, 5},
	}
	got := SumMapSliceLength(mSlice)
	want := 5
	assert.Equal(t, want, got)
}
