package easy

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jxskiss/gopkg/v2/unsafe/reflectx"
	"github.com/jxskiss/gopkg/v2/utils/ptr"
)

type simple struct {
	A string
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

func TestDiff(t *testing.T) {
	s1 := []int{1, 2, 3, 4, 5, 6}
	s2 := []int{2, 4, 6, 8, 10}

	got1 := Diff(s1, s2)
	assert.Equal(t, []int{1, 3, 5}, got1)
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6}, s1)
	assert.Equal(t, []int{2, 4, 6, 8, 10}, s2)

	got2 := Diff(s2, s1)
	assert.Equal(t, []int{8, 10}, got2)
	assert.Equal(t, []int{1, 2, 3, 4, 5, 6}, s1)
	assert.Equal(t, []int{2, 4, 6, 8, 10}, s2)

	s3 := []int{1, 3}
	got3 := Diff(s1, s2, s3)
	assert.Equal(t, []int{5}, got3)
}

func TestFilter(t *testing.T) {
	a := &comptyp{I32: 1, Str_p: ptr.String("a")}
	b := &comptyp{I64: 2, Str_p: ptr.String("b")}
	c := &comptyp{I64_p: ptr.Int64(3), Str_p: ptr.String("c")}
	slice := []*comptyp{a, b, c}

	f1 := func(_ int, x *comptyp) bool { return x.Str_p == nil }
	got1 := Filter(f1, slice)

	assert.NotNil(t, got1)
	assert.Len(t, got1, 0)

	f3 := func(_ int, x *comptyp) bool { return ptr.DerefInt64(x.I64_p) == 3 }
	got3 := Filter(f3, slice)
	assert.Len(t, got3, 1)
}

func callFunction(f interface{}, args ...interface{}) interface{} {
	fVal := reflect.ValueOf(f)
	argsVal := make([]reflect.Value, 0, len(args))
	for _, arg := range args {
		argsVal = append(argsVal, reflect.ValueOf(arg))
	}
	outVals := fVal.Call(argsVal)
	if len(outVals) > 0 {
		return outVals[0].Interface()
	}
	return nil
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

var reverseSliceTests = []map[string]interface{}{
	{
		"func":  ReverseInt64s,
		"slice": []int64{1, 2, 3},
		"want":  []int64{3, 2, 1},
	},
	{
		"func":  ReverseInt32s,
		"slice": []int32{1, 2, 3, 4},
		"want":  []int32{4, 3, 2, 1},
	},
	{
		"func":  ReverseStrings,
		"slice": []string{"1", "2", "3"},
		"want":  []string{"3", "2", "1"},
	},
	{
		"func":  Reverse[[]int8, int8],
		"slice": []int8{1, 2, 3, 4},
		"want":  []int8{4, 3, 2, 1},
	},
	{
		"func":  Reverse[[]simple, simple],
		"slice": []simple{{"a"}, {"b"}, {"c"}, {"d"}},
		"want":  []simple{{"d"}, {"c"}, {"b"}, {"a"}},
	},
	{
		"func":  Reverse[[]int, int],
		"slice": []int(nil),
		"want":  []int(nil),
	},
}

func TestReverseSlice(t *testing.T) {
	for _, test := range reverseSliceTests {
		got := callFunction(test["func"], test["slice"], false)
		assert.Equal(t, test["want"], got)
	}
}

var reverseSliceInplaceTests = []map[string]interface{}{
	{
		"func":  ReverseInt64s,
		"slice": []int64{1, 2, 3},
		"want":  []int64{3, 2, 1},
	},
	{
		"func":  ReverseInt32s,
		"slice": []int32{1, 2, 3},
		"want":  []int32{3, 2, 1},
	},
	{
		"func":  ReverseStrings,
		"slice": []string{"1", "2", "3"},
		"want":  []string{"3", "2", "1"},
	},
	{
		"func":  Reverse[[]int8, int8],
		"slice": []int8{1, 2, 3, 4},
		"want":  []int8{4, 3, 2, 1},
	},
	{
		"func":  Reverse[[]simple, simple],
		"slice": []simple{{"a"}, {"b"}, {"c"}, {"d"}},
		"want":  []simple{{"d"}, {"c"}, {"b"}, {"a"}},
	},
	{
		"func":  Reverse[[]int, int],
		"slice": []int(nil),
		"want":  []int(nil),
	},
}

func TestReverseSliceInplace(t *testing.T) {
	for _, test := range reverseSliceInplaceTests {
		got := callFunction(test["func"], test["slice"], true)
		assert.Equal(t, test["want"], got)
		assert.Equal(t, test["want"], test["slice"])
	}
}

var uniqueSliceTests = []map[string]interface{}{
	{
		"func":  UniqueInt64s,
		"slice": []int64{2, 2, 1, 3, 2, 3, 1, 3},
		"want":  []int64{2, 1, 3},
	},
	{
		"func":  UniqueInt32s,
		"slice": []int32{2, 2, 1, 3, 2, 3, 1, 3},
		"want":  []int32{2, 1, 3},
	},
	{
		"func":  UniqueStrings,
		"slice": []string{"2", "2", "1", "3", "2", "3", "1", "3"},
		"want":  []string{"2", "1", "3"},
	},
}

func TestUniqueSlice(t *testing.T) {
	for _, test := range uniqueSliceTests {
		got := callFunction(test["func"], test["slice"], false)
		assert.Equal(t, test["want"], got)
	}
	for _, test := range uniqueSliceTests {
		got := callFunction(test["func"], test["slice"], true)
		assert.Equal(t, test["want"], got)
		n := reflectx.SliceLen(got)
		changed := reflect.ValueOf(test["slice"]).Slice(0, n).Interface()
		assert.Equal(t, test["want"], changed)
	}
}
