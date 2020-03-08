package easy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var int64sSample = Int64s{5, 6, 7, 8}

var int64sMethodTests = []map[string]interface{}{
	{
		"got":  int64sSample.Uint64s_(),
		"want": []uint64{5, 6, 7, 8},
	},
	{
		"got":  int64sSample.Int32s(),
		"want": []int32{5, 6, 7, 8},
	},
	{
		"got":  int64sSample.Uint32s(),
		"want": []uint32{5, 6, 7, 8},
	},
	{
		"got":  int64sSample.Ints_(),
		"want": []int{5, 6, 7, 8},
	},
	{
		"got":  int64sSample.Uints_(),
		"want": []uint{5, 6, 7, 8},
	},
	{
		"got":  int64sSample.ToStrings(),
		"want": Strings{"5", "6", "7", "8"},
	},
	{
		"got": int64sSample.ToMap(),
		"want": map[int64]bool{
			5: true, 6: true, 7: true, 8: true,
		},
	},
	{
		"got": int64sSample.ToStringMap(),
		"want": map[string]bool{
			"5": true, "6": true, "7": true, "8": true,
		},
	},
}

func TestInt64sMethods(t *testing.T) {
	for _, test := range int64sMethodTests {
		assert.Equal(t, test["want"], test["got"])
	}
}

func TestInt64s_Drop(t *testing.T) {
	slice := []int64{0, 1, 0, 2, 0, 0, 3, 4, 0}
	want := []int64{1, 2, 3, 4}
	length := len(slice)

	var got1 []int64
	got1 = Int64s(slice).Drop(0, false)
	assert.Equal(t, want, got1)
	assert.Equal(t, slice[0], int64(0))

	var got2 []int64
	got2 = Int64s(slice).Drop(0, true)
	assert.Equal(t, want, got2)
	assert.Equal(t, want, slice[:len(got2)])
	assert.Len(t, slice, length)
}

var stringsSample = Strings{"5", "6", "7", "8"}

var stringsMethodTests = []map[string]interface{}{
	{
		"got":  stringsSample.ToInt64s(),
		"want": Int64s{5, 6, 7, 8},
	},
	{
		"got": stringsSample.ToMap(),
		"want": map[string]bool{
			"5": true, "6": true, "7": true, "8": true,
		},
	},
}

func TestStrings_Drop(t *testing.T) {
	slice := []string{"", "a", "b", "", "c", ""}
	want := []string{"a", "b", "c"}
	length := len(slice)

	var got1 []string
	got1 = Strings(slice).Drop("", false)
	assert.Equal(t, want, got1)
	assert.Equal(t, slice[0], "")

	var got2 []string
	got2 = Strings(slice).Drop("", true)
	assert.Equal(t, want, got2)
	assert.Equal(t, want, slice[:len(got2)])
	assert.Len(t, slice, length)
}

func TestToInt64s(t *testing.T) {
	tests := []interface{}{
		[]int8{1, 2, 3},
		[]uint8{1, 2, 3},
		[]int16{1, 2, 3},
		[]uint16{1, 2, 3},
		[]int32{1, 2, 3},
		[]uint32{1, 2, 3},
		[]int64{1, 2, 3},
		[]uint64{1, 2, 3},
		[]int{1, 2, 3},
		[]uint{1, 2, 3},
		[]uintptr{1, 2, 3},
		Int64s{1, 2, 3},
		Strings{"1", "2", "3"},
		[]string{"1", "2", "3"},
		[]string{"1", "a", "2", "", "3", "b"},
	}
	want := Int64s{1, 2, 3}
	for _, test := range tests {
		got := ToInt64s_(test)
		assert.Equal(t, want, got)
	}
}

func TestToInt64s_UnsafeCasting_ChangeOriginal(t *testing.T) {
	if intSize == 32 {
		return
	}

	tests := []map[string]interface{}{
		{
			"slice":  []uint64{1, 2, 3},
			"getter": func(x interface{}, i int) int { return int(x.([]uint64)[i]) },
			"caster": func(x Int64s) interface{} { return x.Uint64s_() },
		},
		{
			"slice":  []int{1, 2, 3},
			"getter": func(x interface{}, i int) int { return int(x.([]int)[i]) },
			"caster": func(x Int64s) interface{} { return x.Ints_() },
		},
		{
			"slice":  []uint{1, 2, 3},
			"getter": func(x interface{}, i int) int { return int(x.([]uint)[i]) },
			"caster": func(x Int64s) interface{} { return x.Uints_() },
		},
		{
			"slice":  []uintptr{1, 2, 3},
			"getter": func(x interface{}, i int) int { return int(x.([]uintptr)[i]) },
		},
	}

	for _, test := range tests {
		slice := test["slice"]
		getter := test["getter"].(func(interface{}, int) int)

		ints := ToInt64s_(slice)
		ints[0], ints[1], ints[2] = 6, 7, 8

		assert.Equal(t, 6, getter(slice, 0))
		assert.Equal(t, 7, getter(slice, 1))
		assert.Equal(t, 8, getter(slice, 2))
	}

	for _, test := range tests {
		getter := test["getter"].(func(interface{}, int) int)
		caster, ok := test["caster"].(func(Int64s) interface{})
		if !ok {
			continue
		}
		newSlice := ToInt64s_(test["slice"]).
			Drop(6, true).
			Drop(7, true)
		test["slice"] = caster(newSlice)

		slice := test["slice"]
		assert.Len(t, slice, 1)
		assert.Equal(t, 8, getter(slice, 0))
	}
}

func TestStringsMethods(t *testing.T) {
	for _, test := range stringsMethodTests {
		assert.Equal(t, test["want"], test["got"])
	}
}

func TestStringsToInt64Map(t *testing.T) {
	want := map[int64]bool{5: true, 6: true, 7: true, 8: true}
	got := stringsSample.ToInt64s().ToMap()
	assert.Equal(t, want, got)
}

func Test_bytes(t *testing.T) {
	text := "Hello, 世界"
	assertEqual := func(buf bytes_) {
		assert.Equal(t, text, buf.String())
		assert.Equal(t, []byte(text), buf.Bytes())
	}

	for _, buf := range []interface{}{
		text,
		[]byte(text),
	} {
		assertEqual(Bytes_(buf))
	}

	assert.Panics(t, func() { Bytes_(12345) })
}
