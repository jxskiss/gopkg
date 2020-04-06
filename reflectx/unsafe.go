package reflectx

import (
	"reflect"
	"unsafe"
)

// StringHeader is a safe version of StringHeader used within this package.
type StringHeader struct {
	Data unsafe.Pointer
	Len  int
}

// SliceHeader is a safe version of SliceHeader used within this package.
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

func EFaceOf(ep *interface{}) *eface {
	return (*eface)(unsafe.Pointer(ep))
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

func add(p unsafe.Pointer, x uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x)
}

func ArrayAt(p unsafe.Pointer, i int, elemSize uintptr) unsafe.Pointer {
	return add(p, uintptr(i)*elemSize)
}

func UnpackSlice(slice interface{}) SliceHeader {
	return *(*SliceHeader)(EFaceOf(&slice).Word)
}

func MakeSlice(elemTyp reflect.Type, length, capacity int) (
	iface interface{}, slice SliceHeader, elemRType unsafe.Pointer,
) {
	elemRType = RTypeOf(elemTyp)
	slice = SliceHeader{
		Data: unsafe_NewArray(elemRType, capacity),
		Len:  length,
		Cap:  capacity,
	}
	iface = *(*interface{})(unsafe.Pointer(&eface{
		RType: RTypeOf(reflect.SliceOf(elemTyp)),
		Word:  unsafe.Pointer(&slice),
	}))
	return
}

func TypedMemMove(rtype unsafe.Pointer, dst, src unsafe.Pointer) {
	typedmemmove(rtype, dst, src)
}
