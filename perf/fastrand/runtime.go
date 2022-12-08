package fastrand

import "github.com/jxskiss/gopkg/v2/internal/linkname"

// Fastrand exposes the fastrand function from runtime package.
func Fastrand() uint32 {
	return linkname.Runtime_fastrand()
}

// Fastrand64 exposes the fastrand64 function from runtime package.
func Fastrand64() uint64 {
	return linkname.Runtime_fastrand64()
}

// Fastrandn exposes the fastrandn function from runtime package.
func Fastrandn(n uint32) uint32 {
	return linkname.Runtime_fastrandn(n)
}

func makeSeed() (ret uint64) {
	ret = linkname.Runtime_fastrand64()
	for ret == 0 {
		ret = linkname.Runtime_fastrand64()
	}
	return
}
