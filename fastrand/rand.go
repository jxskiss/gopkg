// Package fastrand exposes the fastrand functions in runtime package.
package fastrand

import "unsafe"

var _ unsafe.Pointer

//go:noescape
//go:linkname fastrand runtime.fastrand
func fastrand() uint32

//go:noescape
//go:linkname fastrandn runtime.fastrandn
func fastrandn(x uint32) uint32

// Uint32 exposes the fastrand function from runtime package.
func Uint32() uint32 {
	return fastrand()
}

// Uint32n exposes the fastrandn function from runtime package.
func Uint32n(x uint32) uint32 {
	return fastrandn(x)
}
