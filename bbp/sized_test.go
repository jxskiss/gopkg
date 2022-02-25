package bbp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	_4K  = 4096
	_10M = 10 << 20
)

func TestGet(t *testing.T) {
	buf := Get(_4K, _4K)
	t.Log(cap(buf.buf))

	buf = Get(_10M, _10M)
	t.Log(cap(buf.buf))
}

func Test_indexGet(t *testing.T) {
	assert.Equal(t, 6, indexGet(63))
	assert.Equal(t, 6, indexGet(64))
	assert.Equal(t, 7, indexGet(65))
	assert.Equal(t, 7, indexGet(127))
	assert.Equal(t, 7, indexGet(128))
	assert.Equal(t, 8, indexGet(129))
}

func Test_indexPut(t *testing.T) {
	assert.Equal(t, 5, indexPut(63))
	assert.Equal(t, 6, indexPut(64))
	assert.Equal(t, 6, indexPut(65))
	assert.Equal(t, 6, indexPut(127))
	assert.Equal(t, 7, indexPut(128))
	assert.Equal(t, 7, indexPut(129))
	assert.Equal(t, 7, indexPut(255))
	assert.Equal(t, 8, indexPut(256))
}

func BenchmarkAlloc_4K(b *testing.B) {
	var buf []byte
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf = make([]byte, 0, _4K)
	}
	_ = buf
}

func BenchmarkPool_4K(b *testing.B) {
	var buf *Buffer
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf = Get(_4K, _4K)
		Put(buf)
	}
	_ = buf
}

func BenchmarkAlloc_10M(b *testing.B) {
	var buf []byte
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf = make([]byte, 0, _10M)
	}
	_ = buf
}

func BenchmarkPool_10M(b *testing.B) {
	var buf *Buffer
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		buf = Get(_10M, _10M)
		Put(buf)
	}
	_ = buf
}

func BenchmarkAlloc_4K_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		var buf []byte
		for pb.Next() {
			for i := 0; i < 100; i++ {
				buf = make([]byte, 0, _4K)
			}
		}
		_ = buf
	})
}

func BenchmarkPool_4K_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		var buf *Buffer
		for pb.Next() {
			for i := 0; i < 100; i++ {
				buf = Get(_4K, _4K)
				Put(buf)
			}
		}
		_ = buf
	})
}

func BenchmarkAlloc_10M_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		var buf []byte
		for pb.Next() {
			for i := 0; i < 100; i++ {
				buf = make([]byte, 0, _10M)
			}
		}
		_ = buf
	})
}

func BenchmarkPool_10M_Parallel(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		var buf *Buffer
		for pb.Next() {
			for i := 0; i < 100; i++ {
				buf = Get(_10M, _10M)
				Put(buf)
			}
		}
		_ = buf
	})
}
