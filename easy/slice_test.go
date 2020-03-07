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
		"slice": []string{"1", "2", "3", "4"},
		"elem":  "9",
		"index": 3,
		"want":  Strings{"1", "2", "3", "9", "4"},
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
}

func TestPluckStrings(t *testing.T) {
	want := Strings{"a", "b", "c"}
	slice1 := []simple{{"a"}, {"b"}, {"c"}}
	slice2 := []*simple{{"a"}, {"b"}, {"c"}}

	assert.Equal(t, want, PluckStrings(slice1, "A"))
	assert.Equal(t, want, PluckStrings(slice2, "A"))
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
