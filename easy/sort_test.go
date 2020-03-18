package easy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInSortedInt64s(t *testing.T) {
	slice1 := []int64{3, 5, 7, 9, 10}
	slice2 := []int64{10, 9, 7, 5, 3}

	tests := []map[string]interface{}{
		{"elem": 7, "want": true},
		{"elem": 8, "want": false},
		{"elem": 3, "want": true},
		{"elem": 10, "want": true},
		{"elem": 1, "want": false},
		{"elem": 50, "want": false},
	}
	for _, test := range tests {
		got64 := InSortedInt64s(slice1, int64(test["elem"].(int)))
		assert.Equal(t, test["want"], got64)

		got32 := InSortedInt32s(Int64s(slice1).Int32s(), int32(test["elem"].(int)))
		assert.Equal(t, test["want"], got32)

		gotInt := InSortedInts(Int64s(slice1).Ints_(), test["elem"].(int))
		assert.Equal(t, test["want"], gotInt)
	}
	for _, test := range tests {
		got64 := InSortedInt64s(slice2, int64(test["elem"].(int)))
		assert.Equal(t, test["want"], got64)

		got32 := InSortedInt32s(Int64s(slice2).Int32s(), int32(test["elem"].(int)))
		assert.Equal(t, test["want"], got32)

		gotInt := InSortedInts(Int64s(slice2).Ints_(), test["elem"].(int))
		assert.Equal(t, test["want"], gotInt)
	}
}

func TestInSortedStrings(t *testing.T) {
	slice1 := []string{"C", "E", "G", "I", "K"}
	slice2 := []string{"K", "I", "G", "E", "C"}

	tests := []map[string]interface{}{
		{"elem": "G", "want": true},
		{"elem": "H", "want": false},
		{"elem": "C", "want": true},
		{"elem": "K", "want": true},
		{"elem": "A", "want": false},
		{"elem": "Z", "want": false},
	}
	for _, test := range tests {
		got := InSortedStrings(slice1, test["elem"].(string))
		assert.Equal(t, test["want"], got)
	}
	for _, test := range tests {
		got := InSortedStrings(slice2, test["elem"].(string))
		assert.Equal(t, test["want"], got)
	}
}
