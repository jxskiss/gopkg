package lru

import (
	"math/rand"
	"runtime"
	"testing"
	"time"
)

func createFilledShardedCache(ttl time.Duration) *ShardedCache[int64, int64] {
	c := NewShardedCache[int64, int64](8, 500)
	for i := 0; i < 1000; i++ {
		key := int64(rand.Intn(5000))
		c.Set(key, key, ttl)
	}
	return c
}

func TestShardedBasicEviction(t *testing.T) {
	t.Parallel()
	c := NewShardedCache[string, string](4, 3)
	if _, ok, _ := c.Get("a"); ok {
		t.Error("")
	}

	c.Set("b", "vb", 2*time.Second)
	c.Set("a", "va", time.Second)
	c.Set("c", "vc", 3*time.Second)

	if v, _, _ := c.Get("a"); v != "va" {
		t.Error("va")
	}
	if v, _, _ := c.Get("b"); v != "vb" {
		t.Error("vb")
	}
	if v, _, _ := c.Get("c"); v != "vc" {
		t.Error("vc")
	}

	c.MSet(map[string]string{"h": "vh", "i": "vi"}, time.Second)
	if v, _, _ := c.Get("h"); v != "vh" {
		t.Error("vh")
	}
	if v, _, _ := c.Get("i"); v != "vi" {
		t.Error("vi")
	}

	m := c.MGet("h", "i")
	if m["h"] != "vh" {
		t.Error("expecting MSet and NGet to work")
	}
	if m["i"] != "vi" {
		t.Error("expecting MSet and MGet to work")
	}
}

func TestShardedConcurrentGet(t *testing.T) {
	t.Parallel()
	c := createFilledShardedCache(time.Second)
	s := createRandInts(50000)

	runConcurrentGetTest(t, c, s)
}

func TestShardedConcurrentSet(t *testing.T) {
	t.Parallel()
	c := createFilledShardedCache(time.Second)
	s := createRandInts(5000)

	runConcurrentSetTest(t, c, s)
}

func TestShardedConcurrentGetSet(t *testing.T) {
	t.Parallel()
	c := createFilledShardedCache(time.Second)
	s := createRandInts(5000)

	runConcurrentGetSetTest(t, c, s)
}

func BenchmarkShardedConcurrentGetLRUCache(bb *testing.B) {
	c := createFilledShardedCache(time.Second)
	s := createRandInts(5000)

	bb.ReportAllocs()
	bb.ResetTimer()
	cpu := runtime.GOMAXPROCS(0)
	ch := make(chan bool)
	worker := func() {
		for i := 0; i < bb.N/cpu; i++ {
			c.Get(s[i%5000])
		}
		ch <- true
	}
	for i := 0; i < cpu; i++ {
		go worker()
	}
	for i := 0; i < cpu; i++ {
		_ = <-ch
	}
}

func BenchmarkShardedConcurrentSetLRUCache(bb *testing.B) {
	c := createFilledShardedCache(time.Second)
	s := createRandInts(5000)

	bb.ReportAllocs()
	bb.ResetTimer()
	cpu := runtime.GOMAXPROCS(0)
	ch := make(chan bool)
	worker := func() {
		ttl := 4 * time.Second
		for i := 0; i < bb.N/cpu; i++ {
			key := s[i%5000]
			c.Set(key, key, ttl)
		}
		ch <- true
	}
	for i := 0; i < cpu; i++ {
		go worker()
	}
	for i := 0; i < cpu; i++ {
		_ = <-ch
	}
}

// No expiry
func BenchmarkShardedConcurrentSetNXLRUCache(bb *testing.B) {
	c := createFilledShardedCache(time.Second)
	s := createRandInts(5000)

	bb.ReportAllocs()
	bb.ResetTimer()
	cpu := runtime.GOMAXPROCS(0)
	ch := make(chan bool)
	worker := func() {
		for i := 0; i < bb.N/cpu; i++ {
			key := s[i%5000]
			c.Set(key, key, 0)
		}
		ch <- true
	}
	for i := 0; i < cpu; i++ {
		go worker()
	}
	for i := 0; i < cpu; i++ {
		_ = <-ch
	}
}
