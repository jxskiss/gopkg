package monkey

import (
	"reflect"
	"syscall"
	"unsafe"
)

type value struct {
	_   uintptr
	ptr unsafe.Pointer
}

func getPtr(v reflect.Value) unsafe.Pointer {
	return (*value)(unsafe.Pointer(&v)).ptr
}

func getCode(target uintptr, length int) []byte {
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: target,
		Len:  length,
		Cap:  length,
	}))
}

func pageStart(ptr uintptr) uintptr {
	return ptr & ^(uintptr(syscall.Getpagesize() - 1))
}

func copy_(buf []byte) []byte {
	dst := make([]byte, len(buf))
	copy(dst, buf)
	return dst
}

//go:linkname runtime_stopTheWorld runtime.stopTheWorld
func runtime_stopTheWorld()

//go:linkname runtime_startTheWorld runtime.startTheWorld
func runtime_startTheWorld()
