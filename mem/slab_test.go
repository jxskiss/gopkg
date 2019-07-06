package mem

import (
	"bytes"
	"encoding/binary"
	"github.com/gobwas/pool/pbytes"
	"reflect"
	"sort"
	"sync"
	"testing"
)

func Test_Slab_Bytes(t *testing.T) {
	var slab Slab
	var wg sync.WaitGroup

	want := make([]byte, 1000)
	s := "1234567890"
	for i := 0; i < 100; i++ {
		l := len(s)
		copy(want[i*l:(i+1)*l], s)
		wg.Add(1)
		go func() {
			b := slab.Bytes(l)
			b = append(b, s...)
			wg.Done()
		}()
	}
	wg.Wait()

	p := (*block)(slab.buf).p
	if p != 1000 {
		t.Errorf("Slab_Bytes: bad pointer position of slab block")
	}

	b := (*block)(slab.buf).b[:1000]
	if !bytes.Equal(want, b) {
		t.Errorf("Slab_Bytes: bad data of slab block:\nwant = %v\ngot = %v", string(want), string(b))
	}
}

func Test_Slab_Int64(t *testing.T) {
	var slab Slab
	var wg sync.WaitGroup

	want := make([]int, 0, 200)
	for i := 0; i < 100; i++ {
		want = append(want, i, i+1)
		wg.Add(1)
		go func(x int) {
			b := slab.Int64(2)
			b = append(b, int64(x), int64(x+1))
			wg.Done()
		}(i)
	}
	wg.Wait()

	p := (*block)(slab.buf).p
	if p != 1600 {
		t.Errorf("Slab_Int64: bad pointer position of slab block: want = %v got = %v", 1600, p)
	}

	b := (*block)(slab.buf).b[:1600]
	got := make([]int, 0, 200)
	for i := 0; i < 200; i++ {
		got = append(got, int(binary.LittleEndian.Uint64(b[i*8:(i+1)*8])))
	}
	sort.Ints(got)
	if !reflect.DeepEqual(want, got) {
		t.Errorf("Slab_Int64: bad data of slab block:\nwant = %v\ngot = %v", want, got)
	}
}

func Benchmark_SlabBytes(b *testing.B) {
	b.ReportAllocs()
	var slab Slab
	var buf []byte
	for i := 0; i < b.N; i++ {
		buf = slab.Bytes(23)
		_ = buf
	}
}

func Benchmark_GoAlloc(b *testing.B) {
	b.ReportAllocs()
	var buf []byte
	for i := 0; i < b.N; i++ {
		buf = make([]byte, 0, 23)
		_ = buf
	}
}

var syncpool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 23)
	},
}

func Benchmark_SyncPool(b *testing.B) {
	b.ReportAllocs()
	var buf []byte
	for i := 0; i < b.N; i++ {
		buf = syncpool.Get().([]byte)
		_ = buf
		syncpool.Put(buf)
	}
}

var pbytespool = pbytes.New(16, 256)

func Benchmark_PbytesPool(b *testing.B) {
	b.ReportAllocs()
	var buf []byte
	for i := 0; i < b.N; i++ {
		buf = pbytespool.GetCap(23)
		_ = buf
		pbytespool.Put(buf)
	}
}
