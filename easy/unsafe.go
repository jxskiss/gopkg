package easy

import (
	"reflect"
	"unsafe"
)

//go:linkname unsafe_New reflect.unsafe_New
func unsafe_New(unsafe.Pointer) unsafe.Pointer

//go:linkname unsafe_NewArray reflect.unsafe_NewArray
func unsafe_NewArray(unsafe.Pointer, int) unsafe.Pointer

// typedmemmove copies a value of type t to dst from src.
//go:noescape
//go:linkname typedmemmove reflect.typedmemmove
func typedmemmove(t unsafe.Pointer, dst, src unsafe.Pointer)

//go:noescape
//go:linkname maplen reflect.maplen
func maplen(m unsafe.Pointer) int

// m escapes into the return value, but the caller of mapiterinit
// doesn't let the return value escape.
//go:noescape
//go:linkname mapiterinit reflect.mapiterinit
func mapiterinit(rtype unsafe.Pointer, m unsafe.Pointer) *hiter

//go:noescape
//go:linkname mapiternext reflect.mapiternext
func mapiternext(it *hiter)

// A hash iteration structure.
// If you modify hiter, also change cmd/internal/gc/reflect.go to indicate
// the layout of this structure.
type hiter struct {
	key   unsafe.Pointer // Must be in first position.  Write nil to indicate iteration end (see cmd/internal/gc/range.go).
	value unsafe.Pointer // Must be in second position (see cmd/internal/gc/range.go).
	// rest fields are ignored
}

// eface is the header for an interface{} value.
type eface struct {
	typ  unsafe.Pointer // *rtype
	word unsafe.Pointer
}

// stringHeader is a safe version of StringHeader used within this package.
type stringHeader struct {
	Data unsafe.Pointer
	Len  int
}

// sliceHeader is a safe version of SliceHeader used within this package.
type sliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

func efaceOf(ep *interface{}) *eface {
	return (*eface)(unsafe.Pointer(ep))
}

func unpackSlice(slice interface{}) sliceHeader {
	return *(*sliceHeader)(efaceOf(&slice).word)
}

func _iterIntKeys(kind reflect.Kind, m interface{}) []int64 {
	eface := efaceOf(&m)
	hiter := mapiterinit(eface.typ, eface.word)
	tab := int64table[kind]

	keys := make([]int64, 0, maplen(eface.word))
	for hiter.key != nil {
		x := tab.fn(hiter.key)
		keys = append(keys, x)
		mapiternext(hiter)
	}
	return keys
}

func _iterIntValues(kind reflect.Kind, m interface{}) []int64 {
	eface := efaceOf(&m)
	hiter := mapiterinit(eface.typ, eface.word)
	tab := int64table[kind]

	values := make([]int64, 0, maplen(eface.word))
	for hiter.key != nil {
		x := tab.fn(hiter.value)
		values = append(values, x)
		mapiternext(hiter)
	}
	return values
}

func _iterStringKeys(m interface{}) []string {
	eface := efaceOf(&m)
	hiter := mapiterinit(eface.typ, eface.word)

	keys := make([]string, 0, maplen(eface.word))
	for hiter.key != nil {
		x := *(*string)(hiter.key)
		keys = append(keys, x)
		mapiternext(hiter)
	}
	return keys
}

func _iterStringValues(m interface{}) []string {
	eface := efaceOf(&m)
	hiter := mapiterinit(eface.typ, eface.word)

	values := make([]string, 0, maplen(eface.word))
	for hiter.key != nil {
		x := *(*string)(hiter.value)
		values = append(values, x)
		mapiternext(hiter)
	}
	return values
}

func _iterMapKeys_unsafe(m interface{}) interface{} {
	eface := efaceOf(&m)
	hiter := mapiterinit(eface.typ, eface.word)
	length := maplen(eface.word)

	keyTyp := reflect.TypeOf(m).Key()
	keySize := keyTyp.Size()
	out, slice, keyRType := makeSlice(keyTyp, length)
	array := slice.Data
	for i := 0; hiter.key != nil; i++ {
		dst := arrayAt(array, i, keySize)
		typedmemmove(keyRType, dst, hiter.key)
		mapiternext(hiter)
	}
	return out
}

func _iterMapValues_unsafe(m interface{}) interface{} {
	eface := efaceOf(&m)
	hiter := mapiterinit(eface.typ, eface.word)
	length := maplen(eface.word)

	elemTyp := reflect.TypeOf(m).Elem()
	elemSize := elemTyp.Size()
	out, slice, elemRType := makeSlice(elemTyp, length)
	array := slice.Data
	for i := 0; hiter.key != nil; i++ {
		dst := arrayAt(array, i, elemSize)
		typedmemmove(elemRType, dst, hiter.value)
		mapiternext(hiter)
	}
	return out
}

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func s2b(s string) []byte {
	sh := (*stringHeader)(unsafe.Pointer(&s))
	bh := &sliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(bh))
}

func _int32(x interface{}) int32 { return *(*int32)(efaceOf(&x).word) }

func _int64(x interface{}) int64 { return *(*int64)(efaceOf(&x).word) }

func _string(x interface{}) string { return *(*string)(efaceOf(&x).word) }

func rtypeOf(typ reflect.Type) unsafe.Pointer {
	var i interface{} = typ
	return efaceOf(&i).word
}

type i64conv struct {
	sz uintptr
	fn func(unsafe.Pointer) int64
}

var int64table = func() [16]i64conv {
	var table [16]i64conv
	table[reflect.Int8] = i64conv{1, func(p unsafe.Pointer) int64 { return int64(*(*int8)(p)) }}
	table[reflect.Uint8] = i64conv{1, func(p unsafe.Pointer) int64 { return int64(*(*uint8)(p)) }}
	table[reflect.Int16] = i64conv{2, func(p unsafe.Pointer) int64 { return int64(*(*int16)(p)) }}
	table[reflect.Uint16] = i64conv{2, func(p unsafe.Pointer) int64 { return int64(*(*uint16)(p)) }}
	table[reflect.Int32] = i64conv{4, func(p unsafe.Pointer) int64 { return int64(*(*int32)(p)) }}
	table[reflect.Uint32] = i64conv{4, func(p unsafe.Pointer) int64 { return int64(*(*uint32)(p)) }}
	table[reflect.Int64] = i64conv{8, func(p unsafe.Pointer) int64 { return int64(*(*int64)(p)) }}
	table[reflect.Uint64] = i64conv{8, func(p unsafe.Pointer) int64 { return int64(*(*uint64)(p)) }}
	table[reflect.Int] = i64conv{intSize / 8, func(p unsafe.Pointer) int64 { return int64(*(*int)(p)) }}
	table[reflect.Uint] = i64conv{intSize / 8, func(p unsafe.Pointer) int64 { return int64(*(*uint)(p)) }}
	table[reflect.Uintptr] = i64conv{intSize / 8, func(p unsafe.Pointer) int64 { return int64(*(*uintptr)(p)) }}
	return table
}()

func _castInt32s(i32Slice interface{}) []int32 {
	return *(*[]int32)(efaceOf(&i32Slice).word)
}

func _castInt64s(i64Slice interface{}) []int64 {
	return *(*[]int64)(efaceOf(&i64Slice).word)
}

func _convertInt32s(slice interface{}, size uintptr, fn func(unsafe.Pointer) int64) []int32 {
	header := unpackSlice(slice)
	out := make([]int32, header.Len)
	for i := 0; i < header.Len; i++ {
		x := fn(arrayAt(header.Data, i, size))
		out[i] = int32(x)
	}
	return out
}

func _convertInt64s(slice interface{}, size uintptr, fn func(unsafe.Pointer) int64) []int64 {
	header := unpackSlice(slice)
	out := make([]int64, header.Len)
	for i := 0; i < header.Len; i++ {
		x := fn(arrayAt(header.Data, i, size))
		out[i] = x
	}
	return out
}

func arrayAt(p unsafe.Pointer, i int, elemSize uintptr) unsafe.Pointer {
	return add(p, uintptr(i)*elemSize)
}

func add(p unsafe.Pointer, x uintptr) unsafe.Pointer {
	return unsafe.Pointer(uintptr(p) + x)
}

func makeSlice(elemTyp reflect.Type, length int) (
	iface interface{}, slice *sliceHeader, elemRType unsafe.Pointer,
) {
	elemRType = rtypeOf(elemTyp)
	slice = &sliceHeader{
		Data: unsafe_NewArray(elemRType, length),
		Len:  length,
		Cap:  length,
	}
	iface = *(*interface{})(unsafe.Pointer(&eface{
		typ:  rtypeOf(reflect.SliceOf(elemTyp)),
		word: unsafe.Pointer(slice),
	}))
	return
}
