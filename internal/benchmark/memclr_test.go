package benchmark

import (
	"github.com/jxskiss/gopkg/internal/linkname"
	"reflect"
	"testing"
	"unsafe"
)

func BenchmarkMemory_Alloc_4K(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tmp := make([]byte, 4*1024)
		_ = reflect.TypeOf(tmp) // make it escape
	}
}

func BenchmarkMemory_LoopZero_4K(b *testing.B) {
	tmp := make([]byte, 4*1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for j := 0; j < len(tmp); j++ {
			tmp[j] = byte(0)
		}
	}
}

func BenchmarkMemory_memclrNoHeapPointers_4K(b *testing.B) {
	tmp := make([]byte, 4*1024)
	ptr := unsafe.Pointer(&tmp[0])

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		linkname.Runtime_memclrNoHeapPointers(ptr, uintptr(4*1024))
	}
}
