// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package unsafeheader contains header declarations copied from Go's runtime.
package unsafeheader

import (
	"unsafe"
)

// Slice is the runtime representation of a slice.
//
// Unlike reflect.SliceHeader, its Data field is sufficient to guarantee the
// data it references will not be garbage collected.
type Slice struct {
	Data unsafe.Pointer
	Len  int
	Cap  int
}

// String is the runtime representation of a string.
//
// Unlike reflect.StringHeader, its Data field is sufficient to guarantee the
// data it references will not be garbage collected.
type String struct {
	Data unsafe.Pointer
	Len  int
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
