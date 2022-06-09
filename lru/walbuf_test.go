package lru

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestWalbuf(t *testing.T) {
	wb := &walbuf{}

	copy(wb.b[:8], []uint32{3, 1, 3, 4, 9, 1, 10, 10})
	wb.p = 8
	got := wb.deduplicate()
	want := []uint32{3, 4, 9, 1, 10}
	if equal := reflect.DeepEqual(want, got); !equal {
		t.Log(got)
		t.Log(want)
		t.Error("walbuf deduplicate fast path")
	}

	copy(wb.b[:12], []uint32{3, 1, 3, 4, 9, 1, 10, 10, 6, 9, 5, 10})
	wb.p = 12
	got = wb.deduplicate()
	want = []uint32{3, 4, 1, 6, 9, 5, 10}
	if equal := reflect.DeepEqual(want, got); !equal {
		t.Error("walbuf deduplicate slow path")
	}
}

func TestFastHashset(t *testing.T) {
	values := make([]uint32, walBufSize)
	for i := range values {
		values[i] = uint32(rand.Int31n(1000))
	}

	var setBuf [walSetSize]uint32
	fastSet := fastHashset(setBuf)
	for _, x := range values {
		fastSet.add(x)
	}

	fmt.Println(values)
	fmt.Println(setBuf)

	mapSet := make(map[uint32]bool)
	for _, x := range values {
		mapSet[x] = true
	}
	for _, x := range values {
		if fastSet.has(x) != mapSet[x] {
			t.Errorf("got incorrect value from fastHashset")
		}
	}
}

func BenchmarkWalbuf(b *testing.B) {
	cache := NewCache(2000)
	_ = cache

	values := make([]int64, walBufSize)
	for i := range values {
		values[i] = rand.Int63() % walBufSize
	}
	for _, v := range values {
		cache.Set(v, v, 0)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, v := range values {
			_, _, _ = cache.Get(v)
		}
	}
}
