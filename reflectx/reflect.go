package reflectx

import (
	"github.com/jxskiss/gopkg/v2/internal"
	"reflect"
	"unsafe"
)

// IsNilInterface tells whether v is nil or the underlying data is nil.
func IsNilInterface(v interface{}) bool {
	if v == nil {
		return true
	}
	ef := EfaceOf(&v)
	if ef.RType.Kind() == reflect.Slice {
		return *(*unsafe.Pointer)(ef.Word) == nil
	}
	return ef.Word == nil
}

// IsIntType tells whether kind is an integer.
func IsIntType(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true
	}
	return false
}

// ReflectInt returns v's underlying value as int64.
// It panics if v is not a integer value.
func ReflectInt(v reflect.Value) int64 {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return int64(v.Uint())
	}

	// shall not happen, type should be pre-checked
	panic("bug: not int type")
}

// CastInt returns an integer v's value as int64.
// v must be an integer, else it panics.
func CastInt(v interface{}) int64 {
	return internal.CastInt(v)
}

// CastIntPointer returns ptr's value as int64, the underlying value
// is cast to int64 using unsafe tricks according kind.
//
// If ptr is not pointed to an integer or kind does not match ptr,
// the behavior is undefined, it may panic or return incorrect value.
func CastIntPointer(kind reflect.Kind, ptr unsafe.Pointer) int64 {
	return internal.CastIntPointer(kind, ptr)
}
