package easy

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jxskiss/gopkg/v2/utils/ptr"
)

func TestConvInts(t *testing.T) {
	slice1 := []int{1, 2, 3, -3, -2, -1}
	want := []int32{1, 2, 3, -3, -2, -1}
	got := ConvInts[int, int32](slice1)
	assert.Equal(t, want, got)
}

func TestToBoolMap(t *testing.T) {
	slice := []int64{1, 2, 2, 3}
	want := map[int64]bool{1: true, 2: true, 3: true}
	got := ToBoolMap(slice)
	assert.Equal(t, want, got)
}

func TestToMap(t *testing.T) {
	slice := []*comptyp{
		{I32: 2, Simple: simple{"2"}},
		{I32: 3, Simple: simple{"3"}},
		{I32: 3, Simple: simple{"3"}},
	}
	want := map[int32]simple{
		2: {"2"},
		3: {"3"},
	}
	got := ToMap(slice, func(elem *comptyp) (int32, simple) {
		return elem.I32, elem.Simple
	})
	assert.Equal(t, want, got)
}

func TestToInterfaceSlice(t *testing.T) {
	slice1 := []int{1, 2, 3}
	want := []interface{}{1, 2, 3}
	got := ToInterfaceSlice(slice1)
	assert.Equal(t, want, got)

	slice2 := []*int{ptr.Ptr(1), ptr.Ptr(2), ptr.Ptr(3)}
	got2 := ToInterfaceSlice(slice2)
	for i, x := range got2 {
		assert.Equal(t, *slice2[i], *(x.(*int)))
	}

	slice3 := []simple{
		{"a"},
		{"b"},
		{"c"},
	}
	got3 := ToInterfaceSlice(slice3)
	for i, x := range got3 {
		assert.Equal(t, slice3[i], x.(simple))
	}

	slice4 := []*simple{
		{"a"},
		{"b"},
		{"c"},
	}
	got4 := ToInterfaceSlice(slice4)
	for i, x := range got4 {
		assert.Equal(t, slice4[i].A, x.(*simple).A)
	}
}
