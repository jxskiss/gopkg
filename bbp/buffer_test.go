package bbp

import (
	"bytes"
	"testing"
)

func BenchmarkBufferWrite(b *testing.B) {
	s := []byte("foobarbaz")
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		var buf Buffer
		for pb.Next() {
			for i := 0; i < 100; i++ {
				buf.Write(s)
			}
			buf.Reset()
		}
	})
}

func BenchmarkBytesBufferWrite(b *testing.B) {
	s := []byte("foobarbaz")
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		var buf bytes.Buffer
		for pb.Next() {
			for i := 0; i < 100; i++ {
				buf.Write(s)
			}
			buf.Reset()
		}
	})
}
