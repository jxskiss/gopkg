package linkname

import (
	"reflect"
	"unsafe"

	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
)

// ==== exported methods ====

//go:linkname Reflect_rtype_Align reflect.(*rtype).Align
//go:noescape
func Reflect_rtype_Align(unsafe.Pointer) int

//go:linkname Reflect_rtype_FieldAlign reflect.(*rtype).FieldAlign
//go:noescape
func Reflect_rtype_FieldAlign(unsafe.Pointer) int

//go:linkname Reflect_rtype_Method reflect.(*rtype).Method
//go:noescape
func Reflect_rtype_Method(unsafe.Pointer, int) reflect.Method

//go:linkname Reflect_rtype_MethodByName reflect.(*rtype).MethodByName
//go:noescape
func Reflect_rtype_MethodByName(unsafe.Pointer, string) (reflect.Method, bool)

//go:linkname Reflect_rtype_NumMethod reflect.(*rtype).NumMethod
//go:noescape
func Reflect_rtype_NumMethod(unsafe.Pointer) int

//go:linkname Reflect_rtype_Name reflect.(*rtype).Name
//go:noescape
func Reflect_rtype_Name(unsafe.Pointer) string

//go:linkname Reflect_rtype_PkgPath reflect.(*rtype).PkgPath
//go:noescape
func Reflect_rtype_PkgPath(unsafe.Pointer) string

//go:linkname Reflect_rtype_Size reflect.(*rtype).Size
//go:noescape
func Reflect_rtype_Size(unsafe.Pointer) uintptr

//go:linkname Reflect_rtype_String reflect.(*rtype).String
//go:noescape
func Reflect_rtype_String(unsafe.Pointer) string

//go:linkname Reflect_rtype_Kind reflect.(*rtype).Kind
//go:noescape
func Reflect_rtype_Kind(unsafe.Pointer) reflect.Kind

//go:linkname Reflect_rtype_Implements reflect.(*rtype).Implements
//go:noescape
func Reflect_rtype_Implements(unsafe.Pointer, reflect.Type) bool

//go:linkname Reflect_rtype_AssignableTo reflect.(*rtype).AssignableTo
//go:noescape
func Reflect_rtype_AssignableTo(unsafe.Pointer, reflect.Type) bool

//go:linkname Reflect_rtype_ConvertibleTo reflect.(*rtype).ConvertibleTo
//go:noescape
func Reflect_rtype_ConvertibleTo(unsafe.Pointer, reflect.Type) bool

//go:linkname Reflect_rtype_Comparable reflect.(*rtype).Comparable
//go:noescape
func Reflect_rtype_Comparable(unsafe.Pointer) bool

//go:linkname Reflect_rtype_Bits reflect.(*rtype).Bits
//go:noescape
func Reflect_rtype_Bits(unsafe.Pointer) int

//go:linkname Reflect_rtype_ChanDir reflect.(*rtype).ChanDir
//go:noescape
func Reflect_rtype_ChanDir(unsafe.Pointer) reflect.ChanDir

//go:linkname Reflect_rtype_IsVariadic reflect.(*rtype).IsVariadic
//go:noescape
func Reflect_rtype_IsVariadic(unsafe.Pointer) bool

//go:linkname Reflect_rtype_Elem reflect.(*rtype).Elem
//go:noescape
func Reflect_rtype_Elem(unsafe.Pointer) reflect.Type

//go:linkname Reflect_rtype_Field reflect.(*rtype).Field
//go:noescape
func Reflect_rtype_Field(unsafe.Pointer, int) reflect.StructField

//go:linkname Reflect_rtype_FieldByIndex reflect.(*rtype).FieldByIndex
//go:noescape
func Reflect_rtype_FieldByIndex(unsafe.Pointer, []int) reflect.StructField

//go:linkname Reflect_rtype_FieldByName reflect.(*rtype).FieldByName
//go:noescape
func Reflect_rtype_FieldByName(unsafe.Pointer, string) (reflect.StructField, bool)

//go:linkname Reflect_rtype_FieldByNameFunc reflect.(*rtype).FieldByNameFunc
//go:noescape
func Reflect_rtype_FieldByNameFunc(unsafe.Pointer, func(string) bool) (reflect.StructField, bool)

//go:linkname Reflect_rtype_In reflect.(*rtype).In
//go:noescape
func Reflect_rtype_In(unsafe.Pointer, int) reflect.Type

//go:linkname Reflect_rtype_Key reflect.(*rtype).Key
//go:noescape
func Reflect_rtype_Key(unsafe.Pointer) reflect.Type

//go:linkname Reflect_rtype_Len reflect.(*rtype).Len
//go:noescape
func Reflect_rtype_Len(unsafe.Pointer) int

//go:linkname Reflect_rtype_NumField reflect.(*rtype).NumField
//go:noescape
func Reflect_rtype_NumField(unsafe.Pointer) int

//go:linkname Reflect_rtype_NumIn reflect.(*rtype).NumIn
//go:noescape
func Reflect_rtype_NumIn(unsafe.Pointer) int

//go:linkname Reflect_rtype_NumOut reflect.(*rtype).NumOut
//go:noescape
func Reflect_rtype_NumOut(unsafe.Pointer) int

//go:linkname Reflect_rtype_Out reflect.(*rtype).Out
//go:noescape
func Reflect_rtype_Out(unsafe.Pointer, int) reflect.Type

// ==== unexported methods ====

//go:linkname Reflect_typelinks reflect.typelinks
func Reflect_typelinks() ([]unsafe.Pointer, [][]int32)

//go:linkname Reflect_resolveTypeOff reflect.resolveTypeOff
func Reflect_resolveTypeOff(rtype unsafe.Pointer, off int32) unsafe.Pointer

//go:linkname Reflect_rtype_ptrTo reflect.(*rtype).ptrTo
//go:noescape
func Reflect_rtype_ptrTo(unsafe.Pointer) unsafe.Pointer

//go:linkname Reflect_ifaceIndir reflect.ifaceIndir
//go:noescape
func Reflect_ifaceIndir(unsafe.Pointer) bool

//go:linkname Reflect_toType reflect.toType
//go:noescape
func Reflect_toType(unsafe.Pointer) reflect.Type

//go:linkname Reflect_unsafe_New reflect.unsafe_New
func Reflect_unsafe_New(unsafe.Pointer) unsafe.Pointer

//go:linkname Reflect_unsafe_NewArray reflect.unsafe_NewArray
func Reflect_unsafe_NewArray(unsafe.Pointer, int) unsafe.Pointer

// Reflect_typedmemmove copies a value of type t to dst from src.
//
//go:noescape
//go:linkname Reflect_typedmemmove reflect.typedmemmove
func Reflect_typedmemmove(t unsafe.Pointer, dst, src unsafe.Pointer)

// Reflect_typedslicecopy copies a slice of elemType values from src to dst,
// returning the number of elements copied.
//
//go:noescape
//go:linkname Reflect_typedslicecopy reflect.typedslicecopy
func Reflect_typedslicecopy(elemRType unsafe.Pointer, dst, src unsafeheader.Slice) int

//go:noescape
//go:linkname Reflect_maplen reflect.maplen
func Reflect_maplen(m unsafe.Pointer) int

//go:noescape
//go:linkname Reflect_mapiterkey reflect.mapiterkey
func Reflect_mapiterkey(it unsafe.Pointer) (key unsafe.Pointer)

//go:noescape
//go:linkname Reflect_mapiterelem reflect.mapiterelem
func Reflect_mapiterelem(it unsafe.Pointer) (elem unsafe.Pointer)

//go:noescape
//go:linkname Reflect_mapiternext reflect.mapiternext
func Reflect_mapiternext(it unsafe.Pointer)
