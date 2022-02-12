package internal

import (
	"github.com/jxskiss/gopkg/v2/internal/linkname"
	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
	"reflect"
	"unsafe"
)

// EmptyInterface is the header for an interface{} value.
// It's a copy type of runtime.eface.
type EmptyInterface struct {
	RType unsafe.Pointer // *rtype
	Word  unsafe.Pointer
}

// EFaceOf casts the empty interface{} pointer to an EmptyInterface pointer.
func EFaceOf(ep *interface{}) *EmptyInterface {
	return (*EmptyInterface)(unsafe.Pointer(ep))
}

// UnpackSlice unpacks the given slice interface{} to unsafeheader.Slice.
func UnpackSlice(slice interface{}) unsafeheader.Slice {
	return *(*unsafeheader.Slice)(EFaceOf(&slice).Word)
}

// CastInt returns an integer v's value as int64.
// v must be an integer, else it panics.
func CastInt(v interface{}) int64 {
	eface := EFaceOf(&v)
	kind := linkname.Reflect_rtype_Kind(eface.RType)
	return i64table[kind].Cast(eface.Word)
}

// CastIntPointer returns ptr's value as int64, the underlying value
// is cast to int64 using unsafe tricks according kind.
//
// If ptr is not pointed to an integer or kind does not match ptr,
// the behavior is undefined, it may panic or return incorrect value.
func CastIntPointer(kind reflect.Kind, ptr unsafe.Pointer) int64 {
	return i64table[kind].Cast(ptr)
}

const (
	ptrBitSize  = 32 << (^uint(0) >> 63)
	ptrByteSize = ptrBitSize / 8
)

type intInfo struct {
	Size uintptr
	Cast func(unsafe.Pointer) int64
}

var i64table = [...]intInfo{
	reflect.Int8:    {1, func(p unsafe.Pointer) int64 { return int64(*(*int8)(p)) }},
	reflect.Uint8:   {1, func(p unsafe.Pointer) int64 { return int64(*(*uint8)(p)) }},
	reflect.Int16:   {2, func(p unsafe.Pointer) int64 { return int64(*(*int16)(p)) }},
	reflect.Uint16:  {2, func(p unsafe.Pointer) int64 { return int64(*(*uint16)(p)) }},
	reflect.Int32:   {4, func(p unsafe.Pointer) int64 { return int64(*(*int32)(p)) }},
	reflect.Uint32:  {4, func(p unsafe.Pointer) int64 { return int64(*(*uint32)(p)) }},
	reflect.Int64:   {8, func(p unsafe.Pointer) int64 { return int64(*(*int64)(p)) }},
	reflect.Uint64:  {8, func(p unsafe.Pointer) int64 { return int64(*(*uint64)(p)) }},
	reflect.Int:     {ptrByteSize, func(p unsafe.Pointer) int64 { return int64(*(*int)(p)) }},
	reflect.Uint:    {ptrByteSize, func(p unsafe.Pointer) int64 { return int64(*(*uint)(p)) }},
	reflect.Uintptr: {ptrByteSize, func(p unsafe.Pointer) int64 { return int64(*(*uintptr)(p)) }},
}
