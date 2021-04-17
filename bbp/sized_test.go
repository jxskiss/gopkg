package bbp

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	_4K  = 4096
	_10M = 10 << 20
)

func TestGet(t *testing.T) {
	buf := Get(_4K)
	t.Log(cap(buf.B))

	buf = Get(_10M)
	t.Log(cap(buf.B))
}

func Test_indexGet(t *testing.T) {
	assert.Equal(t, 3, indexGet(7))
	assert.Equal(t, 3, indexGet(8))
	assert.Equal(t, 4, indexGet(9))
	assert.Equal(t, 4, indexGet(15))
	assert.Equal(t, 4, indexGet(16))
	assert.Equal(t, 5, indexGet(17))
}

func Test_indexPut(t *testing.T) {
	assert.Equal(t, 2, indexPut(7))
	assert.Equal(t, 3, indexPut(8))
	assert.Equal(t, 3, indexPut(9))
	assert.Equal(t, 3, indexPut(15))
	assert.Equal(t, 4, indexPut(16))
	assert.Equal(t, 4, indexPut(17))
	assert.Equal(t, 4, indexPut(31))
	assert.Equal(t, 5, indexPut(32))
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
		buf = Get(_4K)
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
		buf = Get(_10M)
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
				buf = Get(_4K)
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
				buf = Get(_10M)
				Put(buf)
			}
		}
		_ = buf
	})
}
