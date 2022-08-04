//go:build linux || darwin

package monkey

import (
	"fmt"
	"syscall"
)

const (
	_RW = syscall.PROT_READ | syscall.PROT_WRITE | syscall.PROT_EXEC
	_RX = syscall.PROT_READ | syscall.PROT_EXEC
)

func _replace_code(target uintptr, code []byte) {
	targetCode := getCode(target, len(code))
	_syscall_mprotect(target, len(code), _RW)
	copy(targetCode, code)
	_syscall_mprotect(target, len(code), _RX)
}

func _syscall_mprotect(addr uintptr, size int, prot int) {
	pageSize := syscall.Getpagesize()
	start := pageStart(addr)
	_x := int(addr) + size - int(start)
	pageCount := _x / pageSize
	if _x%pageSize != 0 {
		pageCount += 1
	}
	pages := getCode(start, pageSize*pageCount)
	err := syscall.Mprotect(pages, prot)
	if err != nil {
		panic(fmt.Sprintf("monkey: syscall.Mprotect: %v", err))
	}
}
