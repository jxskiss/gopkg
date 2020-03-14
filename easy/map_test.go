package easy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewSafeMaps(t *testing.T) {
	var _ *SafeMap = NewSafeMap()
	var _ *SafeInt64Map = NewSafeInt64sMap()
	var _ *SafeStringMap = NewSafeStringMap()
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
		Strings{},
		Int64s{},
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
		"keys": Int64s{1, 2, 3},
	},
	{
		"map":  map[int32]int{1: 11, 2: 12, 3: 13},
		"keys": Int64s{1, 2, 3},
	},
	{
		"map":  map[int64]int{1: 11, 2: 13, 3: 13},
		"keys": Int64s{1, 2, 3},
	},
	{
		"map":  map[uint]string{1: "11", 2: "12", 3: "13"},
		"keys": Int64s{1, 2, 3},
	},
	{
		"map": map[uint64]simple{
			1: {"11"}, 2: {"12"}, 3: {"13"},
		},
		"keys": Int64s{1, 2, 3},
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
		"values": Int64s{11, 12, 13},
	},
	{
		"map":    map[int32]uint16{1: 11, 2: 12, 3: 13},
		"values": Int64s{11, 12, 13},
	},
	{
		"map":    map[int64]int64{1: 11, 2: 12, 3: 13},
		"values": Int64s{11, 12, 13},
	},
	{
		"map":    map[string]uint{"1": 11, "2": 12, "3": 13},
		"values": Int64s{11, 12, 13},
	},
	{
		"map": map[simple]int32{
			{"1"}: 11, {"2"}: 12, {"3"}: 13,
		},
		"values": Int64s{11, 12, 13},
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
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		keys := make([]int, 0, len(benchmarkMapData))
		for k := range benchmarkMapData {
			keys = append(keys, k)
		}
		_ = keys
	}
}

func BenchmarkMapKeys_int64s(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = IntKeys(benchmarkMapData)
	}
}

func BenchmarkMapKeys_reflect(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = _iterMapKeys_reflect(benchmarkMapData)
	}
}

func BenchmarkMapKeys_unsafe(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = _iterMapKeys_unsafe(benchmarkMapData)
	}
}

func BenchmarkMapValues_static(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		values := make([]*simple, 0, len(benchmarkMapData))
		for _, v := range benchmarkMapData {
			values = append(values, v)
		}
		_ = values
	}
}

func BenchmarkMapValues_reflect(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = _iterMapValues_reflect(benchmarkMapData)
	}
}

func BenchmarkMapValues_unsafe(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = _iterMapValues_unsafe(benchmarkMapData)
	}
}
