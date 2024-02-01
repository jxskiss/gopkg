//go:build gc && go1.22

package linkname

import _ "unsafe"

func Runtime_fastrand() uint32 {
	return uint32(Runtime_fastrand64())
}

//go:linkname Runtime_fastrand64 runtime.rand
//go:nosplit
func Runtime_fastrand64() uint64
