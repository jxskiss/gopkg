package bbp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestBufferWriteStrings(t *testing.T) {
	strs := []string{
		"hello ",
		"world",
	}
	buf := Get(0, 0)
	buf.WriteStrings(strs)
	want := "hello world"
	assert.Equal(t, 11, buf.Len())
	assert.Equal(t, want, buf.String())
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
