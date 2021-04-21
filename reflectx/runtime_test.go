package reflectx

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type simple struct {
	A string
}

func Test_rtype(t *testing.T) {
	var (
		oneI8    int8  = 1
		twoI32   int32 = 2
		threeInt int   = 3
		strA           = "a"
	)
	types := []interface{}{
		int8(1),
		&oneI8,
		int32(2),
		&twoI32,
		int(3),
		&threeInt,
		"a",
		&strA,
		simple{"b"},
		&simple{"b"},
	}
	for _, x := range types {
		rtype1 := EFaceOf(&x).RType
		rtype2 := RTypeOf(reflect.TypeOf(x))
		rtype3 := RTypeOf(reflect.ValueOf(x))
		assert.Equal(t, rtype1, rtype2)
		assert.Equal(t, rtype2, rtype3)
	}
}

func TestSliceIter(t *testing.T) {
	slice := []int64{1, 2, 3}
	var got []int64
	SliceIter(slice, func(elem interface{}) int {
		got = append(got, elem.(int64))
		return 0
	})
	assert.Equal(t, slice, got)
}
