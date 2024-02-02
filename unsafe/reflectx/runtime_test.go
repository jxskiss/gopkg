package reflectx

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestMakeSlice(t *testing.T) {
	i64Typ := reflect.TypeOf(int64(0))
	_slice, header := MakeSlice(i64Typ, 3, 6)
	assert.NotNil(t, _slice)
	assert.Equal(t, 3, header.Len)
	assert.Equal(t, 6, header.Cap)

	slice := _slice.([]int64)
	slice[0] = 0
	slice[1] = 1
	slice[2] = 2
	slice = append(slice, 3, 4)
	assert.Equal(t, []int64{0, 1, 2, 3, 4}, slice)
	assert.Equal(t, 5, len(slice))
	assert.Equal(t, 6, cap(slice))
	assert.Equal(t, 3, header.Len)
	assert.Equal(t, 6, header.Cap)

	getI64 := func(p unsafe.Pointer) int64 {
		return *(*int64)(p)
	}
	i64Size := unsafe.Sizeof(int64(0))
	assert.Equal(t, int64(0), getI64(ArrayAt(header.Data, 0, i64Size)))
	assert.Equal(t, int64(1), getI64(ArrayAt(header.Data, 1, i64Size)))
	assert.Equal(t, int64(2), getI64(ArrayAt(header.Data, 2, i64Size)))
	assert.Equal(t, int64(3), getI64(ArrayAt(header.Data, 3, i64Size)))
	assert.Equal(t, int64(4), getI64(ArrayAt(header.Data, 4, i64Size)))
	assert.Equal(t, int64(0), getI64(ArrayAt(header.Data, 5, i64Size)))
}

func TestMapLen(t *testing.T) {
	m := make(map[int]bool)
	assert.Equal(t, 0, MapLen(m))

	m[1] = true
	m[2] = false
	assert.Equal(t, 2, MapLen(m))

	type key struct {
		A string
		B [100]byte
		C uint64
	}
	type value struct {
		A string
		B [100]byte
		C uint64
	}

	var dummyBytes [100]byte
	var m2 map[key]*value
	assert.Equal(t, 0, MapLen(m2))
	m2 = make(map[key]*value, 10)
	assert.Equal(t, 0, MapLen(m2))
	m2[key{"abc", dummyBytes, 1234567890}] = &value{"abc", dummyBytes, 1234567890}
	m2[key{"abc", dummyBytes, 1234567890}] = &value{"abc", dummyBytes, 1234567890}
	m2[key{"def", dummyBytes, 1234567890}] = &value{"def", dummyBytes, 1234567890}
	assert.Equal(t, 2, MapLen(m2))

	var m3 map[key]value
	assert.Equal(t, 0, MapLen(m3))
	m3 = make(map[key]value, 100)
	assert.Equal(t, 0, MapLen(m3))
	m3[key{"abc", dummyBytes, 1234567890}] = value{"abc", dummyBytes, 1234567890}
	m3[key{"abc", dummyBytes, 1234567890}] = value{"abc", dummyBytes, 1234567890}
	m3[key{"def", dummyBytes, 1234567890}] = value{"def", dummyBytes, 1234567890}
	assert.Equal(t, 2, MapLen(m3))
}

func TestTypedMemMove(t *testing.T) {
	a := &recurtype2{
		A: "A",
		b: 1,
		self: &recurtype2{
			A:    "AA",
			b:    11,
			self: nil,
		},
	}
	b := &recurtype2{}

	typ := RTypeOf(a)
	TypedMemMove(typ.Elem(), unsafe.Pointer(b), unsafe.Pointer(a))

	assert.Equal(t, a.A, b.A)
	assert.Equal(t, a.b, b.b)
	assert.Equal(t, a.self, b.self)
}

func TestTypedSliceCopy(t *testing.T) {
	slice1 := []string{"a", "b", "c"}
	slice2 := make([]string, 5)

	_, s1header := UnpackSlice(slice1)
	_, s2header := UnpackSlice(slice2)

	elemType := RTypeOf(slice1).Elem()
	TypedSliceCopy(elemType, *s2header, *s1header)

	assert.Equal(t, slice1, slice2[:3])
	assert.Len(t, slice2, 5)
	assert.Equal(t, []string{"", ""}, slice2[3:5])
}
