package unsafeheader

import (
	"reflect"
	"unsafe"
)

func StringToBytes(s string) []byte {
	sh := (*StringHeader)(unsafe.Pointer(&s))
	bh := &SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(bh))
}

func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

// ToEface casts an empty interface{} value to an Eface value.
func ToEface(ep *any) Eface {
	return *(*Eface)(unsafe.Pointer(ep))
}

// ToIface casts a [reflect.Type] to an Iface value.
func ToIface(t reflect.Type) Iface {
	return *(*Iface)(unsafe.Pointer(&t))
}

// ToRType gets the underlying [*reflect.rtype] from a [reflect.Type].
func ToRType(t reflect.Type) unsafe.Pointer {
	return (*Iface)(unsafe.Pointer(&t)).Data
}

// ToReflectType convert an [*reflect.rtype] pointer to a [reflect.Type] value.
// It is the reverse operation of ToRType.
func ToReflectType(rtype unsafe.Pointer) reflect.Type {
	t := reflectTypeTmpl
	t.Data = rtype
	return *(*reflect.Type)(unsafe.Pointer(&t))
}

var reflectTypeTmpl Iface

func init() {
	sampleTyp := reflect.TypeOf(0)
	reflectTypeTmpl = *(*Iface)(unsafe.Pointer(&sampleTyp))
}
