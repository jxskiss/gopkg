//go:build gc && go1.19 && !go1.23

package linkname

import "unsafe"

//go:linkname reflect_ifaceIndir reflect.ifaceIndir
//go:noescape
func reflect_ifaceIndir(rtype unsafe.Pointer) bool
