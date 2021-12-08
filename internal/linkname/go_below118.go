//go:build !go1.18
// +build !go1.18

package linkname

import "unsafe"

// Reflect_mapiterinit .
// m escapes into the return value, but the caller of Reflect_mapiterinit
// doesn't let the return value escape.
//go:noescape
//go:linkname Reflect_mapiterinit reflect.mapiterinit
func Reflect_mapiterinit(rtype unsafe.Pointer, m unsafe.Pointer) unsafe.Pointer
