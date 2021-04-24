package fastrand

import _ "unsafe"

// Fastrand exposes the fastrand function from runtime package.
func Fastrand() uint32 {
	return runtime_fastrand()
}

// Fastrandn exposes the fastrandn function from runtime package.
func Fastrandn(n uint32) uint32 {
	return runtime_fastrandn(n)
}

//go:noescape
//go:linkname runtime_fastrand runtime.fastrand
func runtime_fastrand() uint32

//go:noescape
//go:linkname runtime_fastrandn runtime.fastrandn
func runtime_fastrandn(n uint32) uint32

//go:noescape
//go:linkname runtime_procPin runtime.procPin
func runtime_procPin() int

//go:noescape
//go:linkname runtime_procUnpin runtime.procUnpin
func runtime_procUnpin()
