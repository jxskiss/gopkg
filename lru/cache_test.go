package lru

import (
	"math/rand"
	"runtime"
	"testing"
	"time"
)

func createFilledCache(ttl time.Duration) *Cache {
	c := NewCache(1000)
	for i := 0; i < 1000; i++ {
		key := int64(rand.Intn(5000))
		c.Set(key, key, ttl)
	}
	return c
}

func createRandInts(size int) []int64 {
	s := make([]int64, size)
	for i := 0; i < size; i++ {
		s[i] = rand.Int63n(5000)
	}
	return s
}

func TestBasicEviction(t *testing.T) {
	t.Parallel()
	c := NewCache(3)
	if _, ok, _ := c.Get("a"); ok {
		t.Error("a")
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

	c.Set("d", "vd", time.Second)
	if _, ok, _ := c.Get("a"); ok {
		t.Error("expecting element A to be evicted")
	}
	c.Set("e", "ve", time.Second)
	if _, ok, _ := c.Get("b"); ok {
		t.Error("expecting element B to be evicted")
	}
	c.Set("f", "vf", time.Second)
	if _, ok, _ := c.Get("c"); ok {
		t.Error("expecting element C to be evicted")
	}

	if v, _, _ := c.Get("d"); v != "vd" {
		t.Error("expecting element D to not be evicted")
	}

	// e, f, d, [g]
	c.Set("g", "vg", time.Second)
	if _, ok, _ := c.Get("E"); ok {
		t.Error("expecting element E to be evicted")
	}

	if l := c.Len(); l != 3 {
		t.Errorf("invalid length, want= 3, got= %v", l)
	}

	c.Del("missing")
	c.Del("g")
	if l := c.Len(); l != 2 {
		t.Errorf("invalid length, want= 2, got= %v", l)
	}

	// f, d, [h, i]
	c.MSet(map[string]string{"h": "vh", "i": "vi"}, time.Second)
	if _, ok, _ := c.Get("e"); ok {
		t.Error("expecting element E to be evicted")
	}
	if _, ok, _ := c.Get("f"); ok {
		t.Error("expecting element F to be evicted")
	}
	if v, _, _ := c.Get("d"); v != "vd" {
		t.Error("expecting element D to not be evicted")
	}

	// h/i, i/h, d, [h, i]
	m := c.MGetString("h", "i")
	if m["h"] != "vh" {
		t.Error("expecting MSetString and MGetString to work")
	}
	if m["i"] != "vi" {
		t.Error("expecting MSetString and MGetString to work")
	}

	if v, _, _ := c.GetQuiet("d"); v != "vd" {
		t.Error("expecting GetQuiet to work")
	}

	if v, _ := c.GetNotStale("d"); v != "vd" {
		t.Error("expecting GetNotStale to work")
	}
}

func TestConcurrentGet(t *testing.T) {
	t.Parallel()
	c := createFilledCache(time.Second)
	s := createRandInts(50000)

	done := make(chan bool)
	worker := func() {
		for i := 0; i < 5000; i++ {
			key := s[i]
			v, exists, _ := c.Get(key)
			if exists && v.(int64) != key {
				t.Errorf("value not match: want= %v, got= %v", key, v)
			}
		}
		done <- true
	}
	workers := 4
	for i := 0; i < workers; i++ {
		go worker()
	}
	for i := 0; i < workers; i++ {
		_ = <-done
	}
}

func TestConcurrentSet(t *testing.T) {
	t.Parallel()
	c := createFilledCache(time.Second)
	s := createRandInts(5000)

	done := make(chan bool)
	worker := func() {
		ttl := 4 * time.Second
		for i := 0; i < 5000; i++ {
			key := s[i]
			c.Set(key, key, ttl)
		}
		done <- true
	}
	workers := 4
	for i := 0; i < workers; i++ {
		go worker()
	}
	for i := 0; i < workers; i++ {
		_ = <-done
	}
}

func TestConcurrentGetSet(t *testing.T) {
	t.Parallel()
	c := createFilledCache(time.Second)
	s := createRandInts(5000)

	done := make(chan bool)
	getWorker := func() {
		for i := 0; i < 5000; i++ {
			key := s[i]
			v, exists, _ := c.Get(key)
			if exists && v.(int64) != key {
				t.Errorf("value not match: want= %v, got= %v", key, v)
			}
		}
		done <- true
	}
	setWorker := func() {
		ttl := 4 * time.Second
		for i := 0; i < 5000; i++ {
			key := s[i]
			c.Set(key, key, ttl)
		}
		done <- true
	}
	workers := 4
	for i := 0; i < workers; i++ {
		go getWorker()
		go setWorker()
	}
	for i := 0; i < workers*2; i++ {
		_ = <-done
	}
}

func BenchmarkConcurrentGetLRUCache(bb *testing.B) {
	c := createFilledCache(time.Second)
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

func BenchmarkConcurrentSetLRUCache(bb *testing.B) {
	c := createFilledCache(time.Second)
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
func BenchmarkConcurrentSetNXLRUCache(bb *testing.B) {
	c := createFilledCache(time.Second)
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
