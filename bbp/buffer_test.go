package bbp

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"strings"
	"sync"
	"testing"
)

func TestBufferWrite(t *testing.T) {
	str := "abc"
	want := ""
	buf := NewBuffer(nil)
	for i := 0; i < 1000; i++ {
		tmp := strings.Repeat(str, i)
		want += tmp
		buf.WriteString(tmp)
	}
	got := buf.String()
	assert.Equal(t, want, got)
}

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

var testpool = sync.Pool{
	New: func() interface{} { return &Buffer{} },
}

func BenchmarkSyncPoolWithNew(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = testpool.Get().(*Buffer)
	}
}

func BenchmarkNewBuffer(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = getBuffer()
	}
}
