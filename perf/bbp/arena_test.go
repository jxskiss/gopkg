package bbp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArena(t *testing.T) {
	type arenaIface interface {
		Alloc(length, capacity int) []byte
		Free()
	}

	makeArenas := func() []arenaIface {
		return []arenaIface{
			NewArena(456),
			NewOffHeapArena(567),
		}
	}

	getChunkSize := func(a arenaIface) int {
		switch a := a.(type) {
		case *Arena:
			return a.chunkSize
		case *OffHeapArena:
			return a.chunkSize
		}
		panic("unreachable")
	}

	t.Logf("sysPageSize= %v", sysPageSize)
	for _, a := range makeArenas() {
		assert.Equal(t, sysPageSize, getChunkSize(a))

		n := sysPageSize / 2
		buf := a.Alloc(10, n)
		assert.Equal(t, 10, len(buf))
		assert.Equal(t, n, cap(buf))

		for {
			if n > 2*sysPageSize+1 {
				break
			}
			buf := a.Alloc(10, 100)
			n += cap(buf)
			assert.Equal(t, 10, len(buf))
			assert.Equal(t, 100, cap(buf))
		}
		a.Free()
	}
}

func BenchmarkNewArena(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		a := NewArena(sysPageSize)
		_ = a.Alloc(10, 100)
		a.Free()
	}
}

func BenchmarkNewOffHeapArena(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		a := NewOffHeapArena(sysPageSize)
		_ = a.Alloc(10, 100)
		a.Free()
	}
}
