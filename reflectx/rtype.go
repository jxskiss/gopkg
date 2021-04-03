package reflectx

import (
	"reflect"
	"unsafe"
)

// ---- reflect.Type ---- //

// RType representing reflect.rtype for noescape trick.
// It maps the exported methods of reflect.Type.
type RType struct{}

func (t *RType) Align() int {
	return rtype_Align(t)
}

func (t *RType) FieldAlign() int {
	return rtype_FieldAlign(t)
}

func (t *RType) Method(a0 int) reflect.Method {
	return rtype_Method(t, a0)
}

func (t *RType) MethodByName(a0 string) (reflect.Method, bool) {
	return rtype_MethodByName(t, a0)
}

func (t *RType) NumMethod() int {
	return rtype_NumMethod(t)
}

func (t *RType) Name() string {
	return rtype_Name(t)
}

func (t *RType) PkgPath() string {
	return rtype_PkgPath(t)
}

func (t *RType) Size() uintptr {
	return rtype_Size(t)
}

func (t *RType) String() string {
	return rtype_String(t)
}

func (t *RType) Kind() reflect.Kind {
	return rtype_Kind(t)
}

func (t *RType) Implements(u reflect.Type) bool {
	return rtype_Implements(t, u)
}

func (t *RType) AssignableTo(u reflect.Type) bool {
	return rtype_AssignableTo(t, u)
}

func (t *RType) ConvertibleTo(u reflect.Type) bool {
	return rtype_ConvertibleTo(t, u)
}

func (t *RType) Comparable() bool {
	return rtype_Comparable(t)
}

func (t *RType) Bits() int {
	return rtype_Bits(t)
}

func (t *RType) ChanDir() reflect.ChanDir {
	return rtype_ChanDir(t)
}

func (t *RType) IsVariadic() bool {
	return rtype_IsVariadic(t)
}

func (t *RType) Elem() *RType {
	return ToRType(rtype_Elem(t))
}

func (t *RType) Field(i int) reflect.StructField {
	return rtype_Field(t, i)
}

func (t *RType) FieldByIndex(index []int) reflect.StructField {
	return rtype_FieldByIndex(t, index)
}

func (t *RType) FieldByName(name string) (reflect.StructField, bool) {
	return rtype_FieldByName(t, name)
}

func (t *RType) FieldByNameFunc(match func(string) bool) (reflect.StructField, bool) {
	return rtype_FieldByNameFunc(t, match)
}

func (t *RType) In(i int) reflect.Type {
	return rtype_In(t, i)
}

func (t *RType) Key() *RType {
	return ToRType(rtype_Key(t))
}

func (t *RType) Len() int {
	return rtype_Len(t)
}

func (t *RType) NumField() int {
	return rtype_NumField(t)
}

func (t *RType) NumIn() int {
	return rtype_NumIn(t)
}

func (t *RType) NumOut() int {
	return rtype_NumOut(t)
}

func (t *RType) Out(i int) reflect.Type {
	return rtype_Out(t, i)
}

// ---- extended functions not provided in reflect package ---- //

func (t *RType) IfaceIndir() bool {
	return reflect_ifaceIndir(t)
}

func (t *RType) PackInterface(word unsafe.Pointer) interface{} {
	return *(*interface{})(unsafe.Pointer(&EmptyInterface{
		RType: t,
		Word:  word,
	}))
}

func (t *RType) ReflectType() reflect.Type {
	return reflect_toType(t)
}

func (t *RType) Pointer() unsafe.Pointer {
	return unsafe.Pointer(t)
}

// ---- exported functions ---- //

// PtrTo returns the pointer type with element t.
// For example, if t represents type Foo, PtrTo(t) represents *Foo.
func PtrTo(t *RType) *RType {
	return rtype_ptrTo(t)
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
		return (*value)(unsafe.Pointer(&x)).typ
	default:
		return EFaceOf(&x).RType
	}
}

// ---- below link names to reflect package ---- //

//go:linkname rtype_Align reflect.(*rtype).Align
//go:noescape
func rtype_Align(*RType) int

//go:linkname rtype_FieldAlign reflect.(*rtype).FieldAlign
//go:noescape
func rtype_FieldAlign(*RType) int

//go:linkname rtype_Method reflect.(*rtype).Method
//go:noescape
func rtype_Method(*RType, int) reflect.Method

//go:linkname rtype_MethodByName reflect.(*rtype).MethodByName
//go:noescape
func rtype_MethodByName(*RType, string) (reflect.Method, bool)

//go:linkname rtype_NumMethod reflect.(*rtype).NumMethod
//go:noescape
func rtype_NumMethod(*RType) int

//go:linkname rtype_Name reflect.(*rtype).Name
//go:noescape
func rtype_Name(*RType) string

//go:linkname rtype_PkgPath reflect.(*rtype).PkgPath
//go:noescape
func rtype_PkgPath(*RType) string

//go:linkname rtype_Size reflect.(*rtype).Size
//go:noescape
func rtype_Size(*RType) uintptr

//go:linkname rtype_String reflect.(*rtype).String
//go:noescape
func rtype_String(*RType) string

//go:linkname rtype_Kind reflect.(*rtype).Kind
//go:noescape
func rtype_Kind(*RType) reflect.Kind

//go:linkname rtype_Implements reflect.(*rtype).Implements
//go:noescape
func rtype_Implements(*RType, reflect.Type) bool

//go:linkname rtype_AssignableTo reflect.(*rtype).AssignableTo
//go:noescape
func rtype_AssignableTo(*RType, reflect.Type) bool

//go:linkname rtype_ConvertibleTo reflect.(*rtype).ConvertibleTo
//go:noescape
func rtype_ConvertibleTo(*RType, reflect.Type) bool

//go:linkname rtype_Comparable reflect.(*rtype).Comparable
//go:noescape
func rtype_Comparable(*RType) bool

//go:linkname rtype_Bits reflect.(*rtype).Bits
//go:noescape
func rtype_Bits(*RType) int

//go:linkname rtype_ChanDir reflect.(*rtype).ChanDir
//go:noescape
func rtype_ChanDir(*RType) reflect.ChanDir

//go:linkname rtype_IsVariadic reflect.(*rtype).IsVariadic
//go:noescape
func rtype_IsVariadic(*RType) bool

//go:linkname rtype_Elem reflect.(*rtype).Elem
//go:noescape
func rtype_Elem(*RType) reflect.Type

//go:linkname rtype_Field reflect.(*rtype).Field
//go:noescape
func rtype_Field(*RType, int) reflect.StructField

//go:linkname rtype_FieldByIndex reflect.(*rtype).FieldByIndex
//go:noescape
func rtype_FieldByIndex(*RType, []int) reflect.StructField

//go:linkname rtype_FieldByName reflect.(*rtype).FieldByName
//go:noescape
func rtype_FieldByName(*RType, string) (reflect.StructField, bool)

//go:linkname rtype_FieldByNameFunc reflect.(*rtype).FieldByNameFunc
//go:noescape
func rtype_FieldByNameFunc(*RType, func(string) bool) (reflect.StructField, bool)

//go:linkname rtype_In reflect.(*rtype).In
//go:noescape
func rtype_In(*RType, int) reflect.Type

//go:linkname rtype_Key reflect.(*rtype).Key
//go:noescape
func rtype_Key(*RType) reflect.Type

//go:linkname rtype_Len reflect.(*rtype).Len
//go:noescape
func rtype_Len(*RType) int

//go:linkname rtype_NumField reflect.(*rtype).NumField
//go:noescape
func rtype_NumField(*RType) int

//go:linkname rtype_NumIn reflect.(*rtype).NumIn
//go:noescape
func rtype_NumIn(*RType) int

//go:linkname rtype_NumOut reflect.(*rtype).NumOut
//go:noescape
func rtype_NumOut(*RType) int

//go:linkname rtype_Out reflect.(*rtype).Out
//go:noescape
func rtype_Out(*RType, int) reflect.Type

//go:linkname rtype_ptrTo reflect.(*rtype).ptrTo
//go:noescape
func rtype_ptrTo(*RType) *RType

//go:linkname reflect_ifaceIndir reflect.ifaceIndir
//go:noescape
func reflect_ifaceIndir(*RType) bool

//go:linkname reflect_toType reflect.toType
//go:noescape
func reflect_toType(*RType) reflect.Type
