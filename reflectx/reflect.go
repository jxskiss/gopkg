package reflectx

import (
	"reflect"
	"unsafe"
)

// IsPrivateField returns whether the field is private by checking PkgPath.
//
// PkgPath is the package path that qualifies a lower case (unexported)
// field name. It is empty for upper case (exported) field names.
// See https://golang.org/ref/spec#Uniqueness_of_identifiers
func IsPrivateField(field *reflect.StructField) bool {
	return field.PkgPath != ""
}

// IsIgnoredField returns whether the given field is ignored for specified tag
// by checking the field's anonymity and it's struct tag equals to "-".
func IsIgnoredField(field *reflect.StructField, tag string) bool {
	if field.PkgPath != "" {
		if field.Anonymous {
			if !(field.Type.Kind() == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct) && field.Type.Kind() != reflect.Struct {
				return true
			}
		} else {
			// private field
			return true
		}
	}
	tagVal := field.Tag.Get(tag)
	return tagVal == "-"
}

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
	kind := EFaceOf(&v).RType.Kind()
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
	elemTyp := EFaceOf(&slice).RType.Elem()
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
	elemTyp := EFaceOf(&slice).RType.Elem()
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

var i64table = [...]IntCaster{
	reflect.Int8:    {1, func(p unsafe.Pointer) int64 { return int64(*(*int8)(p)) }},
	reflect.Uint8:   {1, func(p unsafe.Pointer) int64 { return int64(*(*uint8)(p)) }},
	reflect.Int16:   {2, func(p unsafe.Pointer) int64 { return int64(*(*int16)(p)) }},
	reflect.Uint16:  {2, func(p unsafe.Pointer) int64 { return int64(*(*uint16)(p)) }},
	reflect.Int32:   {4, func(p unsafe.Pointer) int64 { return int64(*(*int32)(p)) }},
	reflect.Uint32:  {4, func(p unsafe.Pointer) int64 { return int64(*(*uint32)(p)) }},
	reflect.Int64:   {8, func(p unsafe.Pointer) int64 { return int64(*(*int64)(p)) }},
	reflect.Uint64:  {8, func(p unsafe.Pointer) int64 { return int64(*(*uint64)(p)) }},
	reflect.Int:     {PtrBitSize / 8, func(p unsafe.Pointer) int64 { return int64(*(*int)(p)) }},
	reflect.Uint:    {PtrBitSize / 8, func(p unsafe.Pointer) int64 { return int64(*(*uint)(p)) }},
	reflect.Uintptr: {PtrBitSize / 8, func(p unsafe.Pointer) int64 { return int64(*(*uintptr)(p)) }},
}
