package easy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

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
}

func TestMapKeysValues(t *testing.T) {
	for _, test := range mapKeyValueTests {
		keys := MapKeys(test["map"])
		values := MapValues(test["map"])
		assert.ElementsMatch(t, test["keys"], keys)
		assert.ElementsMatch(t, test["values"], values)
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
		"map":  map[int]string{1: "11", 2: "12", 3: "13"},
		"keys": Int64s{1, 2, 3},
	},
	{
		"map": map[int64]simple{
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
		"map":    map[int32]int16{1: 11, 2: 12, 3: 13},
		"values": Int64s{11, 12, 13},
	},
	{
		"map":    map[int64]int64{1: 11, 2: 12, 3: 13},
		"values": Int64s{11, 12, 13},
	},
	{
		"map":    map[string]int{"1": 11, "2": 12, "3": 13},
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
