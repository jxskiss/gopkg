package easy

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jxskiss/gopkg/v2/unsafe/reflectx"
	"github.com/jxskiss/gopkg/v2/utils/ptr"
)

func TestClip(t *testing.T) {
	s := make([]int, 5, 10)
	got := Clip(s)
	assert.Equal(t, 5, len(got))
	assert.Equal(t, 5, cap(got))
}

func TestCopy(t *testing.T) {
	s := []int64{1, 2, 3}

	got1 := Copy(s)
	assert.Equal(t, got1, s)
	assert.Equal(t, 3, cap(got1))

	got2 := Copy(s, 10)
	assert.Equal(t, got2, s)
	assert.Equal(t, 10, cap(got2))
}

func TestConcat(t *testing.T) {
	s1 := []string{"1", "2"}
	s2 := []string{"a", "b"}
	s3 := []string{"x", "y"}
	want := []string{"1", "2", "a", "b", "x", "y"}
	got := Concat(s1, s2, s3)
	assert.Equal(t, want, got)
}

func TestCount(t *testing.T) {
	s := []string{"1", "2", "a", "b", "3", "4", "x", "y"}
	got := Count(func(s string) bool {
		return s[0] >= '0' && s[0] <= '9'
	}, s)
	assert.Equal(t, 4, got)
}

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

func TestFilterInMap(t *testing.T) {
	s := []int{1, 2, 3, 4, 5}
	m := map[int]string{1: "", 3: "", 5: ""}

	got1 := FilterInMap(s, m, false)
	assert.Equal(t, []int{1, 3, 5}, got1)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, s)

	got2 := FilterInMap(s, m, true)
	assert.Equal(t, []int{1, 3, 5}, got2)
	assert.Equal(t, []int{1, 3, 5, 4, 5}, s)
}

func TestFilterNotInMap(t *testing.T) {
	s := []int{1, 2, 3, 4, 5}
	m := map[int]string{1: "", 3: "", 5: ""}

	got1 := FilterNotInMap(s, m, false)
	assert.Equal(t, []int{2, 4}, got1)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, s)

	got2 := FilterNotInMap(s, m, true)
	assert.Equal(t, []int{2, 4}, got2)
	assert.Equal(t, []int{2, 4, 3, 4, 5}, s)
}

func TestInSlice(t *testing.T) {
	assert.True(t, InSlice([]int32{4, 5, 6}, int32(6)))
	assert.False(t, InSlice([]string{"4", "5", "6"}, "7"))
}

func callFunction(f any, args ...any) any {
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

var splitBatchTests = []map[string]any{
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

func TestRepeat(t *testing.T) {
	s := []int{1, 2, 3}
	got := Repeat(s, 3)
	assert.Equal(t, []int{1, 2, 3, 1, 2, 3, 1, 2, 3}, got)
}

func TestReverseSlice(t *testing.T) {
	var reverseSliceTests = []map[string]any{
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
	for _, test := range reverseSliceTests {
		got := callFunction(test["func"], test["slice"], false)
		assert.Equal(t, test["want"], got)
	}
}

func TestReverseSliceInplace(t *testing.T) {
	var reverseSliceInplaceTests = []map[string]any{
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
	for _, test := range reverseSliceInplaceTests {
		got := callFunction(test["func"], test["slice"], true)
		assert.Equal(t, test["want"], got)
		assert.Equal(t, test["want"], test["slice"])
	}
}

var uniqueSliceTests = []map[string]any{
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

func TestUniqueByLoopCmp(t *testing.T) {
	var dst0 []int64
	src0 := uniqueSliceTests[0]["slice"].([]int64)
	want0 := uniqueSliceTests[0]["want"].([]int64)
	got0 := uniqueByLoopCmp(dst0, src0)
	assert.Equal(t, want0, got0)

	var dst1 []int32
	src1 := uniqueSliceTests[1]["slice"].([]int32)
	want1 := uniqueSliceTests[1]["want"].([]int32)
	got1 := uniqueByLoopCmp(dst1, src1)
	assert.Equal(t, want1, got1)

	var dst2 []string
	src2 := uniqueSliceTests[2]["slice"].([]string)
	want2 := uniqueSliceTests[2]["want"].([]string)
	got2 := uniqueByLoopCmp(dst2, src2)
	assert.Equal(t, want2, got2)
}

func TestUniqueByHashset(t *testing.T) {
	var dst0 []int64
	src0 := uniqueSliceTests[0]["slice"].([]int64)
	want0 := uniqueSliceTests[0]["want"].([]int64)
	got0 := uniqueByHashset(dst0, src0)
	assert.Equal(t, want0, got0)

	var dst1 []int32
	src1 := uniqueSliceTests[1]["slice"].([]int32)
	want1 := uniqueSliceTests[1]["want"].([]int32)
	got1 := uniqueByHashset(dst1, src1)
	assert.Equal(t, want1, got1)

	var dst2 []string
	src2 := uniqueSliceTests[2]["slice"].([]string)
	want2 := uniqueSliceTests[2]["want"].([]string)
	got2 := uniqueByHashset(dst2, src2)
	assert.Equal(t, want2, got2)
}

func TestUniqueFunc(t *testing.T) {
	src0 := uniqueSliceTests[0]["slice"].([]int64)
	want0 := uniqueSliceTests[0]["want"].([]int64)
	got0 := UniqueFunc(src0, false, func(e int64) int32 {
		return int32(e)
	})
	assert.Equal(t, want0, got0)

	src2 := uniqueSliceTests[2]["slice"].([]string)
	want2 := uniqueSliceTests[2]["want"].([]string)
	got2 := UniqueFunc(src2, false, func(e string) string {
		return e
	})
	assert.Equal(t, want2, got2)
}

func TestSum(t *testing.T) {
	s1 := []int{1, 2, 3, 4, 5}
	assert.Equal(t, int64(15), Sum(s1))
	s2 := []uint8{1, 2, 3, 4, 5}
	assert.Equal(t, int64(15), Sum(s2))
}

func TestSumFloat(t *testing.T) {
	s1 := []int{1, 2, 3, 4, 5}
	assert.Equal(t, float64(15), SumFloat(s1))
	s2 := []float32{1, 2, 3, 4, 5}
	assert.Equal(t, float64(15), SumFloat(s2))
}

func TestSort(t *testing.T) {
	s1 := []int{5, 3, 4, 8, 2, 1, 9}
	want1 := []int{1, 2, 3, 4, 5, 8, 9}
	assert.Equal(t, want1, Sort(s1))

	s2 := []string{"b", "c", "a"}
	want2 := []string{"a", "b", "c"}
	assert.Equal(t, want2, Sort(s2))
}

func TestSortDesc(t *testing.T) {
	s1 := []int{5, 3, 4, 8, 2, 1, 9}
	want1 := []int{9, 8, 5, 4, 3, 2, 1}
	assert.Equal(t, want1, SortDesc(s1))

	s2 := []string{"b", "c", "a"}
	want2 := []string{"c", "b", "a"}
	assert.Equal(t, want2, SortDesc(s2))
}

var benchUniqueData []int64
var benchUniqueDst []int64

func initBenchUniqueData() {
	if len(benchUniqueData) > 0 {
		return
	}
	for i := 0; i < 10000; i++ {
		benchUniqueData = append(benchUniqueData, rand.Int63())
	}
	benchUniqueDst = make([]int64, 10000)
}

func BenchmarkUniqueByLoopCmp_64(b *testing.B) {
	initBenchUniqueData()
	f := uniqueByLoopCmp[[]int64, int64]

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got := execUniqueFunc(64, f)
		_ = got
	}
}

func BenchmarkUniqueByHashset_64(b *testing.B) {
	initBenchUniqueData()
	f := uniqueByHashset[[]int64, int64]

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got := execUniqueFunc(64, f)
		_ = got
	}
}

func BenchmarkUniqueByLoopCmp_128(b *testing.B) {
	initBenchUniqueData()
	f := uniqueByLoopCmp[[]int64, int64]

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got := execUniqueFunc(128, f)
		_ = got
	}
}

func BenchmarkUniqueByHashset_128(b *testing.B) {
	initBenchUniqueData()
	f := uniqueByHashset[[]int64, int64]

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got := execUniqueFunc(128, f)
		_ = got
	}
}

func BenchmarkUniqueByLoopCmp_256(b *testing.B) {
	initBenchUniqueData()
	f := uniqueByLoopCmp[[]int64, int64]

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got := execUniqueFunc(256, f)
		_ = got
	}
}

func BenchmarkUniqueByHashset_256(b *testing.B) {
	initBenchUniqueData()
	f := uniqueByHashset[[]int64, int64]

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got := execUniqueFunc(256, f)
		_ = got
	}
}

func BenchmarkUniqueByLoopCmp_512(b *testing.B) {
	initBenchUniqueData()
	f := uniqueByLoopCmp[[]int64, int64]

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got := execUniqueFunc(512, f)
		_ = got
	}
}

func BenchmarkUniqueByHashset_512(b *testing.B) {
	initBenchUniqueData()
	f := uniqueByHashset[[]int64, int64]

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got := execUniqueFunc(512, f)
		_ = got
	}
}

func BenchmarkUniqueByLoopCmp_1024(b *testing.B) {
	initBenchUniqueData()
	f := uniqueByLoopCmp[[]int64, int64]

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got := execUniqueFunc(1024, f)
		_ = got
	}
}

func BenchmarkUniqueByHashset_1024(b *testing.B) {
	initBenchUniqueData()
	f := uniqueByHashset[[]int64, int64]

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		got := execUniqueFunc(1024, f)
		_ = got
	}
}

func execUniqueFunc(length int, f func(dst, src []int64) []int64) []int64 {
	dst := benchUniqueDst[:0]
	src := benchUniqueData[:length]
	return f(dst, src)
}
