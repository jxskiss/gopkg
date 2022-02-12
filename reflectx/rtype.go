package reflectx

import (
	"reflect"

	"github.com/jxskiss/gopkg/v2/internal/rtype"
)

// RType representing reflect.rtype for noescape trick.
// It maps the exported methods of reflect.Type.
type RType = rtype.RType

// ---- exported public functions ---- //

// PtrTo returns the pointer type with element t.
// For example, if t represents type Foo, PtrTo(t) represents *Foo.
func PtrTo(t *RType) *RType {
	return rtype.PtrTo(t)
}

// SliceOf returns the slice type with element type t.
// For example, if t represents int, SliceOf(t) represents []int.
func SliceOf(t *RType) *RType {
	return rtype.SliceOf(t)
}

// MapOf returns the map type with the given key and element types.
// For example, if k represents int and e represents string,
// MapOf(k, e) represents map[int]string.
//
// If the key type is not a valid map key type (that is, if it does
// not implement Go's == operator), MapOf panics.
func MapOf(key, elem *RType) *RType {
	return rtype.MapOf(key, elem)
}

// ToRType converts a reflect.Type value to *RType.
func ToRType(t reflect.Type) *RType {
	return rtype.ToRType(t)
}

// RTypeOf returns the underlying rtype pointer of the given interface{} value.
func RTypeOf(v interface{}) *RType {
	return rtype.RTypeOf(v)
}
