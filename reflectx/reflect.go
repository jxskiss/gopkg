package reflectx

import (
	"reflect"
	"unsafe"
)

func IsStringTypeOrPtr(typ reflect.Type) bool {
	kind := typ.Kind()
	if kind == reflect.Ptr {
		kind = typ.Elem().Kind()
	}
	return kind == reflect.String
}

func IsIntType(kind reflect.Kind) bool {
	switch kind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true
	}
	return false
}

func IsIntTypeOrPtr(typ reflect.Type) bool {
	kind := typ.Kind()
	if kind == reflect.Ptr {
		kind = typ.Elem().Kind()
	}
	return IsIntType(kind)
}

func Is32bitInt(kind reflect.Kind) bool {
	return IsIntType(kind) && GetIntCaster(kind).Size == 4
}

func Is64bitInt(kind reflect.Kind) bool {
	return IsIntType(kind) && GetIntCaster(kind).Size == 8
}

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

func CastInt(v interface{}) int64 {
	kind := reflect.TypeOf(v).Kind()
	return GetIntCaster(kind).Cast(EFaceOf(&v).Word)
}

func CastString(x interface{}) string { return *(*string)(EFaceOf(&x).Word) }

func CastInt32Slice(i32slice interface{}) []int32 {
	return *(*[]int32)(EFaceOf(&i32slice).Word)
}

func CastInt64Slice(i64slice interface{}) []int64 {
	return *(*[]int64)(EFaceOf(&i64slice).Word)
}

func ConvertInt32Slice(slice interface{}) []int32 {
	header := UnpackSlice(slice)
	elemTyp := reflect.TypeOf(slice).Elem()
	info := GetIntCaster(elemTyp.Kind())
	out := make([]int32, header.Len)
	for i := 0; i < header.Len; i++ {
		x := ArrayAt(header.Data, i, info.Size)
		out[i] = int32(info.Cast(x))
	}
	return out
}

func ConvertInt64Slice(slice interface{}) []int64 {
	header := UnpackSlice(slice)
	elemTyp := reflect.TypeOf(slice).Elem()
	info := GetIntCaster(elemTyp.Kind())
	out := make([]int64, header.Len)
	for i := 0; i < header.Len; i++ {
		x := ArrayAt(header.Data, i, info.Size)
		out[i] = info.Cast(x)
	}
	return out
}

type IntCaster struct {
	Size uintptr
	Cast func(unsafe.Pointer) int64
}

func GetIntCaster(kind reflect.Kind) IntCaster {
	return i64table[kind]
}

var i64table [16]IntCaster

func init() {
	i64table[reflect.Int8] = IntCaster{1, func(p unsafe.Pointer) int64 { return int64(*(*int8)(p)) }}
	i64table[reflect.Uint8] = IntCaster{1, func(p unsafe.Pointer) int64 { return int64(*(*uint8)(p)) }}
	i64table[reflect.Int16] = IntCaster{2, func(p unsafe.Pointer) int64 { return int64(*(*int16)(p)) }}
	i64table[reflect.Uint16] = IntCaster{2, func(p unsafe.Pointer) int64 { return int64(*(*uint16)(p)) }}
	i64table[reflect.Int32] = IntCaster{4, func(p unsafe.Pointer) int64 { return int64(*(*int32)(p)) }}
	i64table[reflect.Uint32] = IntCaster{4, func(p unsafe.Pointer) int64 { return int64(*(*uint32)(p)) }}
	i64table[reflect.Int64] = IntCaster{8, func(p unsafe.Pointer) int64 { return int64(*(*int64)(p)) }}
	i64table[reflect.Uint64] = IntCaster{8, func(p unsafe.Pointer) int64 { return int64(*(*uint64)(p)) }}
	i64table[reflect.Int] = IntCaster{IntBitSize / 8, func(p unsafe.Pointer) int64 { return int64(*(*int)(p)) }}
	i64table[reflect.Uint] = IntCaster{IntBitSize / 8, func(p unsafe.Pointer) int64 { return int64(*(*uint)(p)) }}
	i64table[reflect.Uintptr] = IntCaster{IntBitSize / 8, func(p unsafe.Pointer) int64 { return int64(*(*uintptr)(p)) }}
}
