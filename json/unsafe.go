package json

import (
	"github.com/jxskiss/gopkg/reflectx"
	"reflect"
	"unsafe"
)

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func s2b(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := &reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(bh))
}

var strInterfaceMapTyp = reflect.TypeOf(map[string]interface{}(nil))

func isStringInterfaceMap(typ reflect.Type) bool {
	return typ.Kind() == reflect.Map &&
		typ.Key().Kind() == reflect.String &&
		typ.Elem() == strInterfaceMapTyp.Elem()
}

func castStringInterfaceMap(v interface{}) map[string]interface{} {
	eface := reflectx.EFaceOf(&v)
	strMap := *(*map[string]interface{})(unsafe.Pointer(&eface.Word))
	return strMap
}

func isStringMap(typ reflect.Type) bool {
	return typ.Kind() == reflect.Map &&
		typ.Key().Kind() == reflect.String &&
		typ.Elem().Kind() == reflect.String
}

func castStringMap(v interface{}) map[string]string {
	eface := reflectx.EFaceOf(&v)
	strMap := *(*map[string]string)(unsafe.Pointer(&eface.Word))
	return strMap
}

func isStringSlice(typ reflect.Type) bool {
	return typ.Kind() == reflect.Slice &&
		typ.Elem().Kind() == reflect.String
}

func castStringSlice(v interface{}) []string {
	slice := reflectx.UnpackSlice(v)
	return *(*[]string)(unsafe.Pointer(&slice))
}

func isIntSlice(typ reflect.Type) bool {
	return typ.Kind() == reflect.Slice && reflectx.IsIntType(typ.Elem().Kind())
}

func isUnsignedInt(kind reflect.Kind) bool {
	switch kind {
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true
	default:
		return false
	}
}

func castByteSlice(header reflectx.SliceHeader) []byte {
	return *(*[]byte)(unsafe.Pointer(&header))
}

func isStringMapPtr(typ reflect.Type) bool {
	return typ.Kind() == reflect.Ptr && isStringMap(typ.Elem())
}

func castStringMapPtr(v interface{}) *map[string]string {
	eface := reflectx.EFaceOf(&v)
	ptr := (*map[string]string)(eface.Word)
	return ptr
}

func isNilPointer(v interface{}) bool {
	if reflect.TypeOf(v).Kind() != reflect.Ptr {
		return false
	}
	eface := reflectx.EFaceOf(&v)
	return eface.Word == nil
}
