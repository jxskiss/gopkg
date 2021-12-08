package linkname

import (
	"reflect"
	"unsafe"
)

// iface is a copy type of runtime.iface.
type iface struct {
	tab  unsafe.Pointer // *itab
	data unsafe.Pointer
}

// toRType converts a reflect.Type value to *rtype.
func toRType(t reflect.Type) unsafe.Pointer {
	return (*iface)(unsafe.Pointer(&t)).data
}
