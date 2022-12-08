package bbp

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jxskiss/gopkg/v2/perf/fastrand"
)

const (
	_4K  = 4096
	_10M = 10 << 20
)

func TestGet(t *testing.T) {
	buf := Get(_4K, _4K)
	t.Log(cap(buf))

	buf = Get(_10M, _10M)
	t.Log(cap(buf))
}

func Test_indexGet(t *testing.T) {
	for i := 0; i < 10000; i++ {
		size := fastrand.Intn(maxBufSize)
		idx1 := indexGet_readable(size)
		idx2 := indexGet(size)
		assert.Equal(t, idx1, idx2)
	}

	assert.Equal(t, 6, indexGet(63))
	assert.Equal(t, 6, indexGet(64))
	assert.Equal(t, 7, indexGet(65))
	assert.Equal(t, 7, indexGet(127))
	assert.Equal(t, 7, indexGet(128))
	assert.Equal(t, 8, indexGet(129))

	// 14.36 KB
	size := 14701
	idx1 := indexGet_readable(size)
	idx2 := indexGet(size)
	assert.Equal(t, idx1, idx2)
	assert.Equal(t, idx16KB, idx1)
}

func Test_indexPut(t *testing.T) {
	for i := 0; i < 10000; i++ {
		size := fastrand.Intn(maxBufSize)
		idx1 := indexPut_readable(size)
		idx2 := indexPut(size)
		assert.Equalf(t, idx1, idx2, "i= %v, size= %v", i, size)
	}

	assert.Equal(t, 5, indexPut(63))
	assert.Equal(t, 6, indexPut(64))
	assert.Equal(t, 6, indexPut(65))
	assert.Equal(t, 6, indexPut(127))
	assert.Equal(t, 7, indexPut(128))
	assert.Equal(t, 7, indexPut(129))
	assert.Equal(t, 7, indexPut(255))
	assert.Equal(t, 8, indexPut(256))

	// 14.36 KB
	size := 14701
	idx1 := indexPut_readable(size)
	idx2 := indexPut(size)
	assert.Equal(t, idx1, idx2)
	assert.Equal(t, idx12KB, idx1)
}

func Test_indexGet_quarters(t *testing.T) {
	sizeList := []int{
		1139, 1280, 2048, 2049, 4095, 4096, 4097,
		20220, 20480, 28670, 32768,
		262143, 262144, 262145, 393215, 393216, 393217,
	}
	want := []int{
		11, 11, 11, 12, 12, 12, 13,
		16, 16, 18, 19,
		31, 31, 32, 33, 33, 34,
	}
	for i, size := range sizeList {
		assert.Equal(t, want[i], indexGet(size))
	}
}

func Test_indexPut_quarters(t *testing.T) {
	sizeList := []int{
		1139, 1280, 2048, 2049, 4095, 4096, 4097,
		20220, 20480, 28670, 32768,
		262143, 262144, 262145, 393215, 393216, 393217,
	}
	want := []int{
		10, 10, 11, 11, 11, 12, 12,
		15, 16, 17, 19,
		30, 31, 31, 32, 33, 33,
	}
	for i, size := range sizeList {
		assert.Equal(t, want[i], indexPut(size))
	}
}

func Test_various_sizes(t *testing.T) {
	for i := 0; i <= 6; i++ {
		size := 1 << i
		assert.Equal(t, 64, cap(Get(0, size-1)))
		assert.Equal(t, 64, cap(Get(0, size)))
	}
	for i := 7; i <= 12; i++ {
		size := 1 << i
		assert.Equal(t, size, cap(Get(0, size-1)))
		assert.Equal(t, size, cap(Get(0, size)))
		assert.Equal(t, size*2, cap(Get(0, size+1)))
	}
	for i := 13; i <= 25; i++ {
		size := 1 << i
		assert.Equal(t, size, cap(Get(0, size-1)))
		assert.Equal(t, size, cap(Get(0, size)))
		if i < 14 {
			assert.Equal(t, size+size/2, cap(Get(0, size+1)))
		} else if i < 25 {
			assert.Equal(t, size+size/4, cap(Get(0, size+1)))
		} else {
			assert.Equal(t, size+1, cap(Get(0, size+1)))
		}
	}
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
	var buf []byte
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
	var buf []byte
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
		var buf []byte
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
		var buf []byte
		for pb.Next() {
			for i := 0; i < 100; i++ {
				buf = Get(_10M, _10M)
				Put(buf)
			}
		}
		_ = buf
	})
}
