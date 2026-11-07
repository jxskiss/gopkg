package reflectx

import (
	"reflect"
	"unsafe"
)

// IsNil tells whether v is nil or holds a nil value of a nil-capable type.
func IsNil(v any) bool {
	if v == nil {
		return true
	}
	ef := EfaceOf(&v)
	switch ef.RType.Kind() {
	case reflect.Slice:
		return *(*unsafe.Pointer)(ef.Word) == nil
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Pointer, reflect.UnsafePointer:
		return ef.Word == nil
	}
	return false
}

// IsIntType tells whether kind is an integer.
func IsIntType(kind reflect.Kind) (isInt, isSigned bool) {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true, true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true, false
	}
	return false, false
}

// ReflectInt returns v's underlying value as int64.
// It panics if v is not an integer value.
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
