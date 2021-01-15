package reflectx

import (
	"reflect"
	"unsafe"
)

// StringHeader is the runtime representation of a string.
// Unlike reflect.StringHeader, its Data field is sufficient to guarantee the
// data it references will not be garbage collected.
type StringHeader struct {
	Data unsafe.Pointer
	Len  int
}

// SliceHeader is the runtime representation of a slice.
// Unlike reflect.SliceHeader, its Data field is sufficient to guarantee the
// data it references will not be garbage collected.
type SliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func s2b(s string) []byte {
	sh := (*StringHeader)(unsafe.Pointer(&s))
	bh := &SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(bh))
}

func EFaceOf(ep *interface{}) *emptyInterface {
	return (*emptyInterface)(unsafe.Pointer(ep))
}

func PackInterface(typ reflect.Type, word unsafe.Pointer) interface{} {
	var i interface{} = typ
	rtype := EFaceOf(&i).Word
	return *(*interface{})(unsafe.Pointer(&emptyInterface{
		RType: rtype,
		Word:  word,
	}))
}

func RTypeOf(v interface{}) unsafe.Pointer {
	switch v := v.(type) {
	case reflect.Type:
		var i interface{} = v
		return EFaceOf(&i).Word
	case reflect.Value:
		var i interface{} = v.Type()
		return EFaceOf(&i).Word
	default:
		return EFaceOf(&v).RType
	}
}

func MapLen(m interface{}) int {
	return maplen(EFaceOf(&m).Word)
}

func MapIter(m interface{}, f func(k, v unsafe.Pointer)) {
	eface := EFaceOf(&m)
	hiter := mapiterinit(eface.RType, eface.Word)
	for hiter.key != nil {
		f(hiter.key, hiter.value)
		mapiternext(hiter)
	}
}

func add(p unsafe.Pointer, offset uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + offset)
}

// ArrayAt returns the i-th element of p,
// an array whose elements are elemSize bytes wide.
// The array pointed at by p must have at least i+1 elements:
// it is invalid (but impossible to check here) to pass i >= len,
// because then the result will point outside the array.
func ArrayAt(p unsafe.Pointer, i int, elemSize uintptr) unsafe.Pointer {
	return add(p, uintptr(i)*elemSize)
}

func UnpackSlice(slice interface{}) SliceHeader {
	return *(*SliceHeader)(EFaceOf(&slice).Word)
}

func CastSlice(slice interface{}, typ reflect.Type) interface{} {
	newslice := emptyInterface{
		RType: RTypeOf(typ),
		Word:  EFaceOf(&slice).Word,
	}
	return *(*interface{})(unsafe.Pointer(&newslice))
}

func MakeSlice(elemTyp reflect.Type, length, capacity int) (
	slice interface{}, header SliceHeader, elemRType unsafe.Pointer,
) {
	elemRType = RTypeOf(elemTyp)
	header = SliceHeader{
		Data: unsafe_NewArray(elemRType, capacity),
		Len:  length,
		Cap:  capacity,
	}
	slice = *(*interface{})(unsafe.Pointer(&emptyInterface{
		RType: RTypeOf(reflect.SliceOf(elemTyp)),
		Word:  unsafe.Pointer(&header),
	}))
	return
}

func TypedMemMove(rtype unsafe.Pointer, dst, src unsafe.Pointer) {
	typedmemmove(rtype, dst, src)
}
