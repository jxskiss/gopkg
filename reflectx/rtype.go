package reflectx

import (
	"github.com/jxskiss/gopkg/v2/internal"
	"github.com/jxskiss/gopkg/v2/internal/linkname"
	"reflect"
	"unsafe"
)

// ---- reflect.Type ---- //

// RType representing reflect.rtype for noescape trick.
// It maps the exported methods of reflect.Type.
type RType struct{}

func (t *RType) Align() int {
	return linkname.Reflect_rtype_Align(unsafe.Pointer(t))
}

func (t *RType) FieldAlign() int {
	return linkname.Reflect_rtype_FieldAlign(unsafe.Pointer(t))
}

func (t *RType) Method(a0 int) reflect.Method {
	return linkname.Reflect_rtype_Method(unsafe.Pointer(t), a0)
}

func (t *RType) MethodByName(a0 string) (reflect.Method, bool) {
	return linkname.Reflect_rtype_MethodByName(unsafe.Pointer(t), a0)
}

func (t *RType) NumMethod() int {
	return linkname.Reflect_rtype_NumMethod(unsafe.Pointer(t))
}

func (t *RType) Name() string {
	return linkname.Reflect_rtype_Name(unsafe.Pointer(t))
}

func (t *RType) PkgPath() string {
	return linkname.Reflect_rtype_PkgPath(unsafe.Pointer(t))
}

func (t *RType) Size() uintptr {
	return linkname.Reflect_rtype_Size(unsafe.Pointer(t))
}

func (t *RType) String() string {
	return linkname.Reflect_rtype_String(unsafe.Pointer(t))
}

func (t *RType) Kind() reflect.Kind {
	return linkname.Reflect_rtype_Kind(unsafe.Pointer(t))
}

func (t *RType) Implements(u reflect.Type) bool {
	return linkname.Reflect_rtype_Implements(unsafe.Pointer(t), u)
}

func (t *RType) AssignableTo(u reflect.Type) bool {
	return linkname.Reflect_rtype_AssignableTo(unsafe.Pointer(t), u)
}

func (t *RType) ConvertibleTo(u reflect.Type) bool {
	return linkname.Reflect_rtype_ConvertibleTo(unsafe.Pointer(t), u)
}

func (t *RType) Comparable() bool {
	return linkname.Reflect_rtype_Comparable(unsafe.Pointer(t))
}

func (t *RType) Bits() int {
	return linkname.Reflect_rtype_Bits(unsafe.Pointer(t))
}

func (t *RType) ChanDir() reflect.ChanDir {
	return linkname.Reflect_rtype_ChanDir(unsafe.Pointer(t))
}

func (t *RType) IsVariadic() bool {
	return linkname.Reflect_rtype_IsVariadic(unsafe.Pointer(t))
}

func (t *RType) Elem() *RType {
	return ToRType(linkname.Reflect_rtype_Elem(unsafe.Pointer(t)))
}

func (t *RType) Field(i int) reflect.StructField {
	return linkname.Reflect_rtype_Field(unsafe.Pointer(t), i)
}

func (t *RType) FieldByIndex(index []int) reflect.StructField {
	return linkname.Reflect_rtype_FieldByIndex(unsafe.Pointer(t), index)
}

func (t *RType) FieldByName(name string) (reflect.StructField, bool) {
	return linkname.Reflect_rtype_FieldByName(unsafe.Pointer(t), name)
}

func (t *RType) FieldByNameFunc(match func(string) bool) (reflect.StructField, bool) {
	return linkname.Reflect_rtype_FieldByNameFunc(unsafe.Pointer(t), match)
}

func (t *RType) In(i int) reflect.Type {
	return linkname.Reflect_rtype_In(unsafe.Pointer(t), i)
}

func (t *RType) Key() *RType {
	return ToRType(linkname.Reflect_rtype_Key(unsafe.Pointer(t)))
}

func (t *RType) Len() int {
	return linkname.Reflect_rtype_Len(unsafe.Pointer(t))
}

func (t *RType) NumField() int {
	return linkname.Reflect_rtype_NumField(unsafe.Pointer(t))
}

func (t *RType) NumIn() int {
	return linkname.Reflect_rtype_NumIn(unsafe.Pointer(t))
}

func (t *RType) NumOut() int {
	return linkname.Reflect_rtype_NumOut(unsafe.Pointer(t))
}

func (t *RType) Out(i int) reflect.Type {
	return linkname.Reflect_rtype_Out(unsafe.Pointer(t), i)
}

// ---- extended methods not in reflect package ---- //

func (t *RType) IfaceIndir() bool {
	return linkname.Reflect_ifaceIndir(unsafe.Pointer(t))
}

func (t *RType) PackInterface(word unsafe.Pointer) interface{} {
	return *(*interface{})(unsafe.Pointer(&internal.EmptyInterface{
		RType: unsafe.Pointer(t),
		Word:  word,
	}))
}

func (t *RType) ToType() reflect.Type {
	return linkname.Reflect_toType(unsafe.Pointer(t))
}

func (t *RType) Pointer() unsafe.Pointer {
	return unsafe.Pointer(t)
}

// ---- exported public functions ---- //

// PtrTo returns the pointer type with element t.
// For example, if t represents type Foo, PtrTo(t) represents *Foo.
func PtrTo(t *RType) *RType {
	return (*RType)(linkname.Reflect_rtype_ptrTo(unsafe.Pointer(t)))
}

// SliceOf returns the slice type with element type t.
// For example, if t represents int, SliceOf(t) represents []int.
func SliceOf(t *RType) *RType {
	return ToRType(reflect.SliceOf(t.ToType()))
}

// MapOf returns the map type with the given key and element types.
// For example, if k represents int and e represents string,
// MapOf(k, e) represents map[int]string.
//
// If the key type is not a valid map key type (that is, if it does
// not implement Go's == operator), MapOf panics.
func MapOf(key, elem *RType) *RType {
	return ToRType(reflect.MapOf(key.ToType(), elem.ToType()))
}

// ToRType converts a reflect.Type value to *Type.
func ToRType(t reflect.Type) *RType {
	return (*RType)((*iface)(unsafe.Pointer(&t)).data)
}

// RTypeOf returns the underlying rtype pointer of the given interface{} value.
func RTypeOf(v interface{}) *RType {
	switch x := v.(type) {
	case *RType:
		return x
	case reflect.Type:
		return ToRType(x)
	case reflect.Value:
		return ToRType(x.Type())
	default:
		eface := internal.EFaceOf(&x)
		return (*RType)(eface.RType)
	}
}

// ---- private things ---- //

// iface is a copy type of runtime.iface.
type iface struct {
	tab  unsafe.Pointer // *itab
	data unsafe.Pointer
}
