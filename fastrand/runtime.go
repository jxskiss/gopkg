package fastrand

import (
	"github.com/jxskiss/gopkg/internal/linkname"
	_ "unsafe"
)

// Fastrand exposes the fastrand function from runtime package.
func Fastrand() uint32 {
	return linkname.Runtime_fastrand()
}

// Fastrandn exposes the fastrandn function from runtime package.
func Fastrandn(n uint32) uint32 {
	return linkname.Runtime_fastrandn(n)
}
