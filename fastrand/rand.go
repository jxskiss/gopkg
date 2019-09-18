package fastrand

import "unsafe"

var _ unsafe.Pointer

//go:noescape
//go:linkname fastrand runtime.fastrand
func fastrand() uint32

//go:noescape
//go:linkname fastrandn runtime.fastrandn
func fastrandn(x uint32) uint32

func Uint32() uint32 {
	return fastrand()
}

func Uint32n(x uint32) uint32 {
	return fastrandn(x)
}
