package reflectx

import (
	"fmt"
	"github.com/jxskiss/gopkg/internal/unsafeheader"
	"reflect"
	"unsafe"
)

// StringHeader is the runtime representation of a string.
//
// Unlike reflect.StringHeader, its Data field is sufficient to guarantee the
// data it references will not be garbage collected.
type StringHeader = unsafeheader.String

// SliceHeader is the runtime representation of a slice.
//
// Unlike reflect.SliceHeader, its Data field is sufficient to guarantee the
// data it references will not be garbage collected.
type SliceHeader = unsafeheader.Slice

// EmptyInterface is the header for an interface{} value.
// It's a copy type of runtime.eface.
type EmptyInterface struct {
	RType *RType         // *rtype
	Word  unsafe.Pointer // data pointer
}

// StoB converts a string to []byte without copying memory.
//
// It uses unsafe tricks, it may panic your program or result
// unpredictable behavior.
func StoB(s string) []byte {
	return unsafeheader.StoB(s)
}

// BtoS converts a []byte to string without copying memory.
//
// It uses unsafe tricks, it may panic your program or result
// unpredictable behavior.
func BtoS(b []byte) string {
	return unsafeheader.BtoS(b)
}

// EfaceOf casts the empty interface{} pointer to an EmptyInterface pointer.
func EfaceOf(ep *interface{}) *EmptyInterface {
	return (*EmptyInterface)(unsafe.Pointer(ep))
}

// UnpackSlice unpacks the given slice interface{} to the underlying
// EmptyInterface and SliceHeader.
// It panics if param slice is not a slice.
func UnpackSlice(slice interface{}) (*EmptyInterface, *SliceHeader) {
	eface := EfaceOf(&slice)
	if eface.RType.Kind() != reflect.Slice {
		panic(invalidType("UnpackSlice", "slice", slice))
	}
	header := (*SliceHeader)(eface.Word)
	return eface, header
}

// SliceLen returns the length of the given slice interface{} value.
// The provided slice must be a slice, else it panics.
func SliceLen(slice interface{}) int {
	_, header := UnpackSlice(slice)
	return header.Len
}

// SliceCap returns the capacity of the given slice interface{} value.
// The provided slice must be a slice, else it panics.
func SliceCap(slice interface{}) int {
	_, header := UnpackSlice(slice)
	return header.Cap
}

func invalidType(where string, want string, got interface{}) string {
	const invalidType = "%s: invalid type, want %s, got %T"
	return fmt.Sprintf(invalidType, where, want, got)
}
