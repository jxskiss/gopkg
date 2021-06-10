package gemap

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type simple struct {
	A string
}

var mapKeyValueTests = []map[string]interface{}{
	{
		"map":    map[int]int{1: 11, 2: 12, 3: 13},
		"keys":   []int{1, 2, 3},
		"values": []int{11, 12, 13},
	},
	{
		"map":    map[string]string{"1": "11", "2": "12", "3": "13"},
		"keys":   []string{"1", "2", "3"},
		"values": []string{"11", "12", "13"},
	},
	{
		"map": map[simple]simple{
			{"1"}: {"11"},
			{"2"}: {"12"},
			{"3"}: {"13"},
		},
		"keys":   []simple{{"1"}, {"2"}, {"3"}},
		"values": []simple{{"11"}, {"12"}, {"13"}},
	},
	{
		"map":    map[int32]string{1: "11", 2: "12", 3: "13"},
		"keys":   []int32{1, 2, 3},
		"values": []string{"11", "12", "13"},
	},
	{
		"map":    map[int64]int64{1: 11, 2: 12, 3: 13},
		"keys":   []int64{1, 2, 3},
		"values": []int64{11, 12, 13},
	},
}

func TestMapKeysValues(t *testing.T) {
	for _, test := range mapKeyValueTests {
		keys := MapKeys(test["map"])
		values := MapValues(test["map"])
		assert.ElementsMatch(t, test["keys"], keys)
		assert.ElementsMatch(t, test["values"], values)
	}
}

func TestKeysValues_Int_String(t *testing.T) {
	intMap := map[uint16]string{1: "a", 2: "b", 3: "c"}
	stringMap := map[string]uint8{"a": 1, "b": 2, "c": 3}
	assert.Panics(t, func() { _ = IntValues(intMap) })
	assert.Panics(t, func() { _ = StringKeys(intMap) })
	assert.Panics(t, func() { _ = IntKeys(stringMap) })
	assert.Panics(t, func() { _ = StringValues(stringMap) })

	intWant := []int64{1, 2, 3}
	strWant := []string{"a", "b", "c"}
	assert.ElementsMatch(t, intWant, IntKeys(intMap))
	assert.ElementsMatch(t, intWant, IntValues(stringMap))
	assert.ElementsMatch(t, strWant, StringValues(intMap))
	assert.ElementsMatch(t, strWant, StringKeys(stringMap))
}

func TestMapKeysValues_panic(t *testing.T) {
	notMapTests := []interface{}{
		123,
		[]int{1, 2, 3},
		simple{"a"},
		&simple{"b"},
		[]*simple{{"a"}, {"b"}, {"c"}},
		[]string{},
	}
	for _, test := range notMapTests {
		assert.Panics(t, func() { _ = MapKeys(test) })
		assert.Panics(t, func() { _ = MapValues(test) })
		assert.Panics(t, func() { _ = IntKeys(test) })
		assert.Panics(t, func() { _ = IntValues(test) })
		assert.Panics(t, func() { _ = StringKeys(test) })
		assert.Panics(t, func() { _ = StringValues(test) })
	}
}

var intKeysTests = []map[string]interface{}{
	{
		"map":  map[int]int{1: 11, 2: 12, 3: 13},
		"keys": []int64{1, 2, 3},
	},
	{
		"map":  map[int32]int{1: 11, 2: 12, 3: 13},
		"keys": []int64{1, 2, 3},
	},
	{
		"map":  map[int64]int{1: 11, 2: 13, 3: 13},
		"keys": []int64{1, 2, 3},
	},
	{
		"map":  map[uint]string{1: "11", 2: "12", 3: "13"},
		"keys": []int64{1, 2, 3},
	},
	{
		"map": map[uint64]simple{
			1: {"11"}, 2: {"12"}, 3: {"13"},
		},
		"keys": []int64{1, 2, 3},
	},
}

func TestIntKeys(t *testing.T) {
	for _, test := range intKeysTests {
		got := IntKeys(test["map"])
		assert.ElementsMatch(t, test["keys"], got)
	}
}

var intValuesTests = []map[string]interface{}{
	{
		"map":    map[int]int{1: 11, 2: 12, 3: 13},
		"values": []int64{11, 12, 13},
	},
	{
		"map":    map[int32]uint16{1: 11, 2: 12, 3: 13},
		"values": []int64{11, 12, 13},
	},
	{
		"map":    map[int64]int64{1: 11, 2: 12, 3: 13},
		"values": []int64{11, 12, 13},
	},
	{
		"map":    map[string]uint{"1": 11, "2": 12, "3": 13},
		"values": []int64{11, 12, 13},
	},
	{
		"map": map[simple]int32{
			{"1"}: 11, {"2"}: 12, {"3"}: 13,
		},
		"values": []int64{11, 12, 13},
	},
}

func TestIntValues(t *testing.T) {
	for _, test := range intValuesTests {
		got := IntValues(test["map"])
		assert.ElementsMatch(t, test["values"], got)
	}
}

func TestStringKeysValues(t *testing.T) {
	m := map[string]string{"1": "11", "2": "12", "3": "13"}
	assert.ElementsMatch(t, []string{"1", "2", "3"}, StringKeys(m))
	assert.ElementsMatch(t, []string{"11", "12", "13"}, StringValues(m))
}

func TestMergeMaps(t *testing.T) {
	m1 := map[int64]int64{1: 2, 3: 4, 5: 6}
	m2 := map[int64]int64{7: 8, 9: 10}
	got := MergeMaps(m1, m2).(map[int64]int64)
	assert.Equal(t, 3, len(m1))
	assert.Equal(t, 2, len(m2))
	assert.Equal(t, 5, len(got))
	assert.Equal(t, int64(4), got[3])
	assert.Equal(t, int64(8), got[7])
}

func TestMergeMapsTo(t *testing.T) {
	m1 := map[int64]int64{1: 2, 3: 4, 5: 6}
	m2 := map[int64]int64{7: 8, 9: 10}
	_ = MergeMapsTo(m1, m2).(map[int64]int64)
	assert.Equal(t, 5, len(m1))
	assert.Equal(t, 2, len(m2))
	assert.Equal(t, int64(4), m1[3])
	assert.Equal(t, int64(8), m1[7])
	assert.Equal(t, int64(10), m1[9])

	// merge to a nil map
	var m3 map[int64]int64
	m4 := MergeMapsTo(m3, m1).(map[int64]int64)
	assert.Nil(t, m3)
	assert.Equal(t, 5, len(m4))
	assert.Equal(t, int64(4), m4[3])
	assert.Equal(t, int64(10), m4[9])
}

var benchmarkMapData = map[int]*simple{
	1:  {"abc"},
	2:  {"bcd"},
	3:  {"cde"},
	4:  {"def"},
	5:  {"efg"},
	6:  {"fgh"},
	7:  {"ghi"},
	8:  {"hij"},
	9:  {"ijk"},
	10: {"jkl"},
}

func BenchmarkMapKeys_static(b *testing.B) {

	MapKeys_static := func(m map[int]*simple) []int {
		keys := make([]int, 0, len(benchmarkMapData))
		for k := range benchmarkMapData {
			keys = append(keys, k)
		}
		return keys
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = MapKeys_static(benchmarkMapData)
	}
}

func BenchmarkMapKeys_reflect(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = MapKeys(benchmarkMapData)
	}
}

func BenchmarkMapKeys_int64(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = IntKeys(benchmarkMapData)
	}
}

func BenchmarkMapValues_static(b *testing.B) {

	MapValues_static := func(m map[int]*simple) []*simple {
		values := make([]*simple, 0, len(benchmarkMapData))
		for _, v := range benchmarkMapData {
			values = append(values, v)
		}
		return values
	}

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = MapValues_static(benchmarkMapData)
	}
}

func BenchmarkMapValues_reflect(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = MapValues(benchmarkMapData)
	}
}
