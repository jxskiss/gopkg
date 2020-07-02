package reflectx

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

var convertInt32SliceTestCases = []map[string]interface{}{
	{
		"slice": []int8{-1, 0, 1, math.MaxInt8},
		"want":  []int32{-1, 0, 1, math.MaxInt8},
	},
	{
		"slice": []uint32{0, 1, math.MaxUint32},
		"want":  []int32{0, 1, -1},
	},
	{
		"slice": []uint64{0, 1, math.MaxUint64},
		"want":  []int32{0, 1, -1},
	},
}

func TestConvertInt32Slice(t *testing.T) {
	for _, test := range convertInt32SliceTestCases {
		got := ConvertInt32Slice(test["slice"])
		want := test["want"]
		assert.Equal(t, want, got)
	}
}

var convertInt64SliceTestCases = []map[string]interface{}{
	{
		"slice": []int8{-1, 0, 1, math.MaxInt8},
		"want":  []int64{-1, 0, 1, math.MaxInt8},
	},
	{
		"slice": []uint32{0, 1, math.MaxUint32},
		"want":  []int64{0, 1, math.MaxUint32},
	},
	{
		"slice": []uint64{0, 1, math.MaxUint64},
		"want":  []int64{0, 1, -1},
	},
}

func TestConvertInt64Slice(t *testing.T) {
	for _, test := range convertInt64SliceTestCases {
		got := ConvertInt64Slice(test["slice"])
		want := test["want"]
		assert.Equal(t, want, got)
	}
}
