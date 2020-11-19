// Package fastrand exposes the fastrand functions in runtime package.
package fastrand

import _ "unsafe"

// Fastrand exposes the fastrand function from runtime package.
func Fastrand() uint32 {
	return runtime_fastrand()
}

// Fastrandn exposes the fastrandn function from runtime package.
func Fastrandn(x uint32) uint32 {
	return runtime_fastrandn(x)
}

//go:noescape
//go:linkname fastrand runtime.fastrand
func runtime_fastrand() uint32

//go:noescape
//go:linkname fastrandn runtime.fastrandn
func runtime_fastrandn(x uint32) uint32
