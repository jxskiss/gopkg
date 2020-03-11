package easy

import (
	"github.com/jxskiss/gopkg/ptr"
	"github.com/stretchr/testify/assert"
	"testing"
)

type simple struct {
	A string
}

var inSliceTests = []map[string]interface{}{
	{
		"func":  InSlice,
		"slice": []int64{1, 2, 3, 4},
		"elem":  int64(3),
		"want":  true,
	},
	{
		"func":  InSlice,
		"slice": []int{1, 2, 3, 4},
		"elem":  3,
		"want":  true,
	},
	{
		"func":  InSlice,
		"slice": []int64{1, 2, 3, 4},
		"elem":  int64(8),
		"want":  false,
	},
	{
		"func":  InSlice,
		"slice": []int64{1, 2, 3, 4},
		"elem":  int(3), // int type not match
		"want":  true,
	},
	{
		"func":  InSlice,
		"slice": []int64{1, 2, 3, 4},
		"elem":  int16(8), // int type not match
		"want":  false,
	},
	{
		"func":  InSlice,
		"slice": []string{"1", "2", "3", "4"},
		"elem":  "3",
		"want":  true,
	},
	{
		"func":  InSlice,
		"slice": []string{"1", "2", "3", "4"},
		"elem":  "a",
		"want":  false,
	},
	{
		"func":  InSlice,
		"slice": Int64s{1, 2, 3, 4},
		"elem":  3,
		"want":  true,
	},
	{
		"func":  InSlice,
		"slice": Strings{"1", "2", "3", "4"},
		"elem":  "3",
		"want":  true,
	},
	{
		"func":  InSlice,
		"slice": []simple{{"a"}, {"b"}, {"c"}, {"d"}},
		"elem":  simple{"c"},
		"want":  true,
	},
	{
		"func":  InSlice,
		"slice": []simple{{"a"}, {"b"}, {"c"}, {"d"}},
		"elem":  simple{"z"},
		"want":  false,
	},
	{
		"func":  InSlice,
		"slice": nil,
		"elem":  "a",
		"want":  false,
	},
	{
		"func":  InSlice,
		"slice": Int64s{1, 2, 3},
		"elem":  nil,
		"want":  false,
	},
}

func TestInSlice(t *testing.T) {
	for _, test := range inSliceTests {
		f := test["func"].(func(slice interface{}, elem interface{}) bool)
		got := f(test["slice"], test["elem"])
		assert.Equal(t, test["want"], got)
	}
}

var insertSliceTests = []map[string]interface{}{
	{
		"func":  InsertSlice,
		"slice": []int64{1, 2, 3, 4},
		"elem":  int64(9),
		"index": 3,
		"want":  Int64s{1, 2, 3, 9, 4},
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
		"want":  Int64s{1, 2, 3, 4, 9},
	},
	{
		"func":  InsertSlice,
		"slice": []string{"1", "2", "3", "4"},
		"elem":  "9",
		"index": 3,
		"want":  Strings{"1", "2", "3", "9", "4"},
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
		"want":  nil,
	},
	{
		"func":  InsertSlice,
		"slice": []int64{1, 2, 3, 4},
		"elem":  nil,
		"index": 3,
		"want":  []int64{1, 2, 3, 4},
	},
}

func TestInsertSlice(t *testing.T) {
	for _, test := range insertSliceTests {
		f := test["func"].(func(slice interface{}, index int, elem interface{}) interface{})
		got := f(test["slice"], test["index"].(int), test["elem"])
		assert.Equal(t, test["want"], got)
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
	assert.Equal(t, want, Pluck(&slice1, "A"))
	assert.Equal(t, want, Pluck(&slice2, "A"))

	assert.Nil(t, Pluck(nil, "A"))
	assert.Nil(t, Pluck(slice1, ""))
}

func TestPluckStrings(t *testing.T) {
	want := Strings{"a", "b", "c"}
	slice1 := []simple{{"a"}, {"b"}, {"c"}}
	slice2 := []*simple{{"a"}, {"b"}, {"c"}}

	assert.Equal(t, want, PluckStrings(slice1, "A"))
	assert.Equal(t, want, PluckStrings(slice2, "A"))
	assert.Equal(t, want, PluckStrings(&slice1, "A"))
	assert.Equal(t, want, PluckStrings(&slice2, "A"))

	assert.Nil(t, PluckStrings(nil, "A"))
	assert.Nil(t, PluckStrings(slice1, ""))
}

func TestPluckInt64s(t *testing.T) {
	slice := complexTypeData

	got1 := PluckInt64s(slice, "I32")
	want1 := Int64s{32, 33, 34, 35}
	assert.Equal(t, want1, got1)
	assert.Equal(t, want1, PluckInt64s(&slice, "I32"))

	got2 := PluckInt64s(slice, "I32_p")
	want2 := Int64s{32, 33, 34}
	assert.Equal(t, want2, got2)
	assert.Equal(t, want2, PluckInt64s(&slice, "I32_p"))

	got3 := PluckInt64s(slice, "I64")
	want3 := Int64s{64, 65, 66, 67}
	assert.Equal(t, want3, got3)
	assert.Equal(t, want3, PluckInt64s(&slice, "I64"))

	got4 := PluckInt64s(slice, "I64_p")
	want4 := Int64s{64, 65, 66}
	assert.Equal(t, want4, got4)
	assert.Equal(t, want4, PluckInt64s(&slice, "I64_p"))

	assert.Nil(t, PluckInt64s(nil, "I32"))
	assert.Nil(t, PluckInt64s(slice, ""))
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

var indexSliceTests = []map[string]interface{}{
	{
		"slice": nil,
		"elem":  2,
		"want":  -1,
	},
	{
		"slice": []int{1, 2, 3},
		"elem":  nil,
		"want":  -1,
	},
	{
		"slice": []int64{1, 2, 3},
		"elem":  int64(2),
		"want":  1,
	},
	{
		"slice": []int64{1, 2, 3},
		"elem":  int32(2), // int type not match
		"want":  1,
	},
	{
		"slice": []uint{1, 2, 3},
		"elem":  2,
		"want":  1,
	},
	{
		"slice": []uint{1, 2, 3},
		"elem":  4,
		"want":  -1,
	},
	{
		"slice": []string{"1", "2", "3"},
		"elem":  "2",
		"want":  1,
	},
	{
		"slice": []string{"1", "2", "3"},
		"elem":  "a",
		"want":  -1,
	},
	{
		"slice": []simple{{"a"}, {"b"}, {"c"}},
		"elem":  simple{"b"},
		"want":  1,
	},
	{
		"slice": []*simple{{"a"}, {"b"}, {"c"}},
		"elem":  &simple{"b"},
		"want":  -1,
	},
}

func TestIndex(t *testing.T) {
	for _, test := range indexSliceTests {
		got := Index(test["slice"], test["elem"])
		assert.Equal(t, test["want"], got)
	}
}

var lastIndexSliceTests = []map[string]interface{}{
	{
		"slice": nil,
		"elem":  2,
		"want":  -1,
	},
	{
		"slice": []int{1, 2, 3, 2},
		"elem":  nil,
		"want":  -1,
	},
	{
		"slice": []int64{1, 2, 3, 2},
		"elem":  int64(2),
		"want":  3,
	},
	{
		"slice": []int64{1, 2, 3, 2},
		"elem":  int32(2), // int type not match
		"want":  3,
	},
	{
		"slice": []string{"1", "2", "3", "2"},
		"elem":  "2",
		"want":  3,
	},
	{
		"slice": []string{"1", "2", "3", "2"},
		"elem":  "a",
		"want":  -1,
	},
	{
		"slice": []simple{{"a"}, {"b"}, {"c"}, {"b"}},
		"elem":  simple{"b"},
		"want":  3,
	},
	{
		"slice": []*simple{{"a"}, {"b"}, {"c"}, {"b"}},
		"elem":  &simple{"b"},
		"want":  -1,
	},
}

func TestLastIndex(t *testing.T) {
	for _, test := range lastIndexSliceTests {
		got := LastIndex(test["slice"], test["elem"])
		assert.Equal(t, test["want"], got)
	}
}

var reverseSliceTests = []map[string]interface{}{
	{
		"slice": nil,
		"want":  nil,
	},
	{
		"slice": []uint64{1, 2, 3},
		"want":  Int64s{3, 2, 1},
	},
	{
		"slice": []int8{1, 2, 3},
		"want":  []int8{3, 2, 1},
	},
	{
		"slice": []string{"1", "2", "3"},
		"want":  Strings{"3", "2", "1"},
	},
	{
		"slice": []simple{{"a"}, {"b"}, {"c"}},
		"want":  []simple{{"c"}, {"b"}, {"a"}},
	},
}

func TestReverseSlice(t *testing.T) {
	for _, test := range reverseSliceTests {
		got := ReverseSlice(test["slice"])
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

	assert.Nil(t, ToMap(nil, "A"))
	assert.Nil(t, ToMap(slice, ""))
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

func TestToInt64Map(t *testing.T) {
	a := &comptyp{I32_p: ptr.Int32(1)}
	b := &comptyp{I32_p: ptr.Int32(2)}
	c := &comptyp{I32_p: ptr.Int32(3)}
	slice := []*comptyp{a, b, c}
	want := map[int64]*comptyp{1: a, 2: b, 3: c}
	got := ToInt64Map(slice, "I32_p")
	assert.Equal(t, want, got)

	assert.Nil(t, ToInt64Map(nil, "I32_p"))
	assert.Nil(t, ToInt64Map(slice, ""))
}

func TestToStringMap(t *testing.T) {
	a := &comptyp{Str_p: ptr.String("a")}
	b := &comptyp{Str_p: ptr.String("b")}
	c := &comptyp{Str_p: ptr.String("c")}
	slice := []*comptyp{a, b, c}
	want := map[string]*comptyp{"a": a, "b": b, "c": c}
	got := ToStringMap(slice, "Str_p")
	assert.Equal(t, want, got)

	assert.Nil(t, ToStringMap(nil, "Str_p"))
	assert.Nil(t, ToStringMap(slice, ""))
}

func TestParseCommaInt64s(t *testing.T) {
	strIntIDs := "123,,456,789, ,0,"
	want := Int64s{123, 456, 789}
	got, isMalformed := ParseCommaInt64s(strIntIDs, true)
	assert.True(t, isMalformed)
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
