package linkname

import (
	"reflect"
	"unsafe"
)

// eface is the header for an interface{} value.
// It's a copy type of [runtime.eface].
type eface struct {
	rtype unsafe.Pointer // *rtype
	data  unsafe.Pointer // data pointer
}

func unpackEface(ep *any) *eface {
	return (*eface)(unsafe.Pointer(ep))
}

// iface is a copy type of [runtime.iface].
type iface struct {
	tab  unsafe.Pointer // *itab
	data unsafe.Pointer
}

// ToRType converts a [reflect.Type] value to [*reflect.rtype].
func ToRType(t reflect.Type) unsafe.Pointer {
	return (*iface)(unsafe.Pointer(&t)).data
}
