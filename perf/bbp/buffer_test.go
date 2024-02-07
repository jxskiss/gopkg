package bbp

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBufferReadFrom(t *testing.T) {
	str := strings.Repeat("hello world", 1<<18)
	buf := NewBuffer(0)
	n, err := buf.ReadFrom(strings.NewReader(str))
	assert.Nil(t, err)
	assert.Equal(t, 11*(1<<18), int(n))
	assert.Equal(t, str, buf.String())
	assert.Equal(t, str, buf.StringUnsafe())
}

func TestBufferWrite(t *testing.T) {
	str := "abc"
	want := ""
	var buf Buffer
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
	for i := 0; i < 18; i++ { // -> 2_883_584
		strs = append(strs, strs...)
	}
	buf := NewBuffer(0)
	buf.WriteStrings(strs)
	want := strings.Repeat("hello world", 1<<18)
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
