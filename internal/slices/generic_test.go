package slices

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type simple struct {
	A string
}

var insertTests = []map[string]interface{}{
	{
		"slice": []int64{1, 2, 3, 4},
		"elem":  int64(9),
		"index": 3,
		"want":  []int64{1, 2, 3, 9, 4},
	},
	{
		"slice": []int32{1, 2, 3, 4},
		"elem":  int32(9),
		"index": 4,
		"want":  []int32{1, 2, 3, 4, 9},
	},
	{
		"slice": []string{"1", "2", "3", "4"},
		"elem":  "9",
		"index": 3,
		"want":  []string{"1", "2", "3", "9", "4"},
	},
}

func TestInsert(t *testing.T) {
	var s0 = insertTests[0]["slice"].([]int64)
	var elem0 = insertTests[0]["elem"].(int64)
	var idx0 = insertTests[0]["index"].(int)
	var want0 = insertTests[0]["want"]
	got0 := Insert(s0, idx0, elem0)
	assert.Equal(t, want0, got0)

	var s1 = insertTests[1]["slice"].([]int32)
	var elem1 = insertTests[1]["elem"].(int32)
	var idx1 = insertTests[1]["index"].(int)
	var want1 = insertTests[1]["want"]
	got1 := Insert(s1, idx1, elem1)
	assert.Equal(t, want1, got1)

	var s2 = insertTests[2]["slice"].([]string)
	var elem2 = insertTests[2]["elem"].(string)
	var idx2 = insertTests[2]["index"].(int)
	var want2 = insertTests[2]["want"]
	got2 := Insert(s2, idx2, elem2)
	assert.Equal(t, want2, got2)
}

var reverseTests = []map[string]interface{}{
	{
		"slice": []int64{1, 2, 3},
		"want":  []int64{3, 2, 1},
	},
	{
		"slice": []string{"1", "2", "3"},
		"want":  []string{"3", "2", "1"},
	},
	{
		"slice": []simple{{"a"}, {"b"}, {"c"}, {"d"}},
		"want":  []simple{{"d"}, {"c"}, {"b"}, {"a"}},
	},
	{
		"slice": []int(nil),
		"want":  []int(nil),
	},
}

func TestReverse(t *testing.T) {
	var s0 = reverseTests[0]["slice"].([]int64)
	var want0 = reverseTests[0]["want"]
	got0 := Reverse(s0, false)
	assert.Equal(t, want0, got0)

	var s1 = reverseTests[1]["slice"].([]string)
	var want1 = reverseTests[1]["want"]
	got1 := Reverse(s1, false)
	assert.Equal(t, want1, got1)

	var s2 = reverseTests[2]["slice"].([]simple)
	var want2 = reverseTests[2]["want"]
	got2 := Reverse(s2, false)
	assert.Equal(t, want2, got2)

	var s3 = reverseTests[3]["slice"].([]int)
	var want3 = reverseTests[3]["want"]
	got3 := Reverse(s3, false)
	assert.Equal(t, want3, got3)
}

var reverseInplaceTests = []map[string]interface{}{
	{
		"slice": []int64{1, 2, 3},
		"want":  []int64{3, 2, 1},
	},
	{
		"slice": []string{"1", "2", "3"},
		"want":  []string{"3", "2", "1"},
	},
	{
		"slice": []simple{{"a"}, {"b"}, {"c"}, {"d"}},
		"want":  []simple{{"d"}, {"c"}, {"b"}, {"a"}},
	},
	{
		"slice": []int(nil),
		"want":  []int(nil),
	},
}

func TestReverse_inplace(t *testing.T) {
	var s0 = reverseInplaceTests[0]["slice"].([]int64)
	var want0 = reverseInplaceTests[0]["want"]
	got0 := Reverse(s0, true)
	assert.Equal(t, want0, got0)
	assert.Equal(t, want0, s0)

	var s1 = reverseInplaceTests[1]["slice"].([]string)
	var want1 = reverseInplaceTests[1]["want"]
	got1 := Reverse(s1, true)
	assert.Equal(t, want1, got1)
	assert.Equal(t, want1, s1)

	var s2 = reverseInplaceTests[2]["slice"].([]simple)
	var want2 = reverseInplaceTests[2]["want"]
	got2 := Reverse(s2, true)
	assert.Equal(t, want2, got2)
	assert.Equal(t, want2, s2)

	var s3 = reverseInplaceTests[3]["slice"].([]int)
	var want3 = reverseInplaceTests[3]["want"]
	got3 := Reverse(s3, true)
	assert.Equal(t, want3, got3)
	assert.Equal(t, want3, s3)
}

var uniqueSliceTests = []map[string]interface{}{
	{
		"slice": []int64{2, 2, 1, 3, 2, 3, 1, 3},
		"want":  []int64{2, 1, 3},
	},
	{
		"slice": []string{"2", "2", "1", "3", "2", "3", "1", "3"},
		"want":  []string{"2", "1", "3"},
	},
}

func TestUnique(t *testing.T) {
	var s0 = uniqueSliceTests[0]["slice"].([]int64)
	var want0 = uniqueSliceTests[0]["want"]
	got0 := Unique(s0, false)
	assert.Equal(t, want0, got0)

	var s1 = uniqueSliceTests[1]["slice"].([]string)
	var want1 = uniqueSliceTests[1]["want"]
	got1 := Unique(s1, false)
	assert.Equal(t, want1, got1)
}

func TestDiff_int64(t *testing.T) {
	a := []int64{1, 2, 3, 4, 5}
	b := []int64{4, 5, 6, 7, 8}
	want1 := []int64{1, 2, 3}
	assert.Equal(t, want1, Diff(a, b))

	want2 := []int64{6, 7, 8}
	assert.Equal(t, want2, Diff(b, a))
}

func TestDiff_string(t *testing.T) {
	a := []string{"1", "2", "3", "4", "5"}
	b := []string{"4", "5", "6", "7", "8"}
	want1 := []string{"1", "2", "3"}
	assert.Equal(t, want1, Diff(a, b))

	want2 := []string{"6", "7", "8"}
	assert.Equal(t, want2, Diff(b, a))
}

func TestSplitSlice(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6, 7}

	testTable := []struct {
		batch int
		want  [][]int
	}{
		{-1, [][]int{{1, 2, 3, 4, 5, 6, 7}}},
		{0, [][]int{{1, 2, 3, 4, 5, 6, 7}}},
		{1, [][]int{{1}, {2}, {3}, {4}, {5}, {6}, {7}}},
		{2, [][]int{{1, 2}, {3, 4}, {5, 6}, {7}}},
		{3, [][]int{{1, 2, 3}, {4, 5, 6}, {7}}},
		{4, [][]int{{1, 2, 3, 4}, {5, 6, 7}}},
		{7, [][]int{{1, 2, 3, 4, 5, 6, 7}}},
		{8, [][]int{{1, 2, 3, 4, 5, 6, 7}}},
	}
	for _, c := range testTable {
		want := c.want
		got := Split(slice, c.batch)
		assert.Equal(t, want, got)
	}
}

func TestSumSlice(t *testing.T) {
	assert.Equal(t, int64(0), Sum([]int(nil)))
	assert.Equal(t, int64(6), Sum([]int32{1, 2, 3}))
	assert.Equal(t, int64(15), Sum([]int64{4, 5, 6}))
}
