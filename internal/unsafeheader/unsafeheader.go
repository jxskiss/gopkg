// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package unsafeheader contains header declarations copied from Go's runtime.
package unsafeheader

import (
	"unsafe"
)

// SliceHeader is the runtime representation of a slice.
//
// Unlike reflect.SliceHeader, its Data field is sufficient to guarantee the
// data it references will not be garbage collected.
type SliceHeader struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

// SliceData returns a pointer to the underlying array of the argument
// slice.
//   - If len(slice) == 0, it returns nil.
//   - Otherwise, it returns the underlying pointer.
func SliceData[T any](slice []T) unsafe.Pointer {
	if len(slice) == 0 {
		return nil
	}
	return (*SliceHeader)(unsafe.Pointer(&slice)).Data
}

// StringHeader is the runtime representation of a string.
//
// Unlike reflect.StringHeader, its Data field is sufficient to guarantee the
// data it references will not be garbage collected.
type StringHeader struct {
	Data unsafe.Pointer
	Len  int
}

// StringData returns a pointer to the underlying bytes of str.
//   - If str == "", it returns nil.
//   - Otherwise, it returns the underlying pointer.
func StringData(str string) unsafe.Pointer {
	if str == "" {
		return nil
	}
	return (*StringHeader)(unsafe.Pointer(&str)).Data
}

// Eface is the header for an empty interface{} value.
// It is a copy type of [runtime.eface].
type Eface struct {
	RType unsafe.Pointer // *rtype
	Word  unsafe.Pointer // data pointer
}

// Iface is the header of a non-empty interface value.
// It is a copy type of [runtime.iface].
type Iface struct {
	Tab  unsafe.Pointer // *itab
	Data unsafe.Pointer
}
