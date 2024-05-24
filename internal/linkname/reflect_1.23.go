//go:build gc && go1.23

package linkname

import "unsafe"

//go:linkname reflect_ifaceIndir internal/abi.(*Type).IfaceIndir
//go:noescape
func reflect_ifaceIndir(rtype unsafe.Pointer) bool
