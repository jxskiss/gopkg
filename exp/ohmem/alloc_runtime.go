//go:build !mimalloc

package ohmem

import "github.com/jxskiss/gopkg/v2/internal/linkname"

func _C_zalloc(n int) []byte {
	mem := linkname.Runtime_sysAlloc(uintptr(n))
	for i := range mem {
		mem[i] = 0
	}
	return mem
}

func _C_free(mem []byte) {
	linkname.Runtime_sysFree(mem)
}
