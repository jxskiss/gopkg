package unsafeheader

import "unsafe"

func StoB(s string) []byte {
	sh := (*String)(unsafe.Pointer(&s))
	bh := &Slice{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(bh))
}

func BtoS(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
