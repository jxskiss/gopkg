package fastrand

import _ "unsafe"

//go:noescape
//go:linkname fastrand runtime.fastrand
func fastrand() uint32

//go:noescape
//go:linkname fastrandn runtime.fastrandn
func fastrandn(x uint32) uint32

//go:linkname runtime_procPin runtime.procPin
func runtime_procPin() int

//go:linkname runtime_procUnpin runtime.procUnpin
func runtime_procUnpin()

//go:nosplit
func procHint() int {
	// More detail discussion can be found here:
	// https://github.com/golang/go/issues/18590

	pid := runtime_procPin()
	runtime_procUnpin()
	return pid
}
