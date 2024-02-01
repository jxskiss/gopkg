//go:build gc && go1.19 && !go1.22

package linkname

import _ "unsafe"

//go:linkname Runtime_fastrand runtime.fastrand
//go:nosplit
func Runtime_fastrand() uint32

//go:linkname Runtime_fastrand64 runtime.fastrand64
//go:nosplit
func Runtime_fastrand64() uint64
