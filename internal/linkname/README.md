Package linkname exports various private functions from the standard library
using the `//go:linkname` directive.

**DON'T USE THIS IF YOU DON'T KNOW WHAT IT IS.**

Also, it is a bad practice to use private functions from other packages,
and it is **UNSAFE and not protected by the Go 1 compatibility promise**.
