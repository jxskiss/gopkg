package fastrand

import (
	"syscall"
	"testing"
)

func BenchmarkSyscallGetpid(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = syscall.Getpid()
	}
}

func BenchmarkProcHint(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = procHint()
	}
}
