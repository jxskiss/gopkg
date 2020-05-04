// Package fastrand exposes the fastrand functions in runtime package.
package fastrand

// Uint32 exposes the fastrand function from runtime package.
func Uint32() uint32 {
	return fastrand()
}

// Uint32n exposes the fastrandn function from runtime package.
func Uint32n(x uint32) uint32 {
	return fastrandn(x)
}
