package monkey

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	_RW = 0x40 // PAGE_EXECUTE_READWRITE
	_RX = 0x20 // PAGE_EXECUTE_READ
)

var virtualProtectProc = syscall.
	NewLazyDLL("kernel32.dll").NewProc("VirtualProtect")

func _replace_code(target uintptr, code []byte) {
	targetCode := getCode(target, len(code))

	var old, ignore uint32
	_virtual_protect(target, len(code), _RW, &old)
	copy(targetCode, code)
	_virtual_protect(target, len(code), old, &ignore)
}

func _virtual_protect(addr uintptr, dwSize int, newProtect uint32, oldProtect *uint32) {
	r1, _, lastErr := virtualProtectProc.Call(
		addr,
		uintptr(dwSize),
		uintptr(newProtect),
		uintptr(unsafe.Pointer(oldProtect)),
	)
	if r1 == 0 {
		panic(fmt.Sprintf("monkey: kernel32.dll.VirtualProtect: %v", lastErr))
	}
}
