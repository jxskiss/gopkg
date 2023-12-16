package reflectx

import (
	"reflect"
	"testing"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestStringToBytes(t *testing.T) {
	s := "hello, world"
	b := StringToBytes(s)
	assert.Equal(t, []byte(s), b)

	strHeader := (*StringHeader)(unsafe.Pointer(&s))
	bytesHeader := (*SliceHeader)(unsafe.Pointer(&b))
	assert.Equal(t, strHeader.Data, bytesHeader.Data)
}

func TestBytesToString(t *testing.T) {
	b := []byte("hello, world")
	s := BytesToString(b)
	assert.Equal(t, "hello, world", s)

	bytesHeader := (*SliceHeader)(unsafe.Pointer(&b))
	strHeader := (*StringHeader)(unsafe.Pointer(&s))
	assert.Equal(t, bytesHeader.Data, strHeader.Data)
}

func TestUnpackSlice(t *testing.T) {
	slice := make([]int, 3, 6)
	slice[0], slice[1], slice[2] = 0, 1, 2
	eface, header := UnpackSlice(slice)

	assert.Equal(t, 3, header.Len)
	assert.Equal(t, 6, header.Cap)
	assert.Equal(t, reflect.Slice, eface.RType.Kind())
	assert.Equal(t, reflect.Int, eface.RType.Elem().Kind())

	assert.Equal(t, 3, SliceLen(slice))
	assert.Equal(t, 6, SliceCap(slice))
}
