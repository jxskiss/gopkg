package lru

import (
	"math/rand"
	"runtime"
	"testing"
	"time"
)

func createFilledMultiCache(ttl time.Duration) *multiCache {
	c := NewMultiCache(8, 500)
	for i := 0; i < 1000; i++ {
		key := int64(rand.Intn(5000))
		c.Set(key, key, ttl)
	}
	return c
}

func TestMultiBasicEviction(t *testing.T) {
	t.Parallel()
	c := NewMultiCache(4, 3)
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

	m := c.MGetString("h", "i")
	if m["h"] != "vh" {
		t.Error("expecting MSetString and MGetString to work")
	}
	if m["i"] != "vi" {
		t.Error("expecting MSetString and MGetString to work")
	}
}

func TestMultiConcurrentGet(t *testing.T) {
	t.Parallel()
	c := createFilledMultiCache(time.Second)
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

func TestMultiConcurrentSet(t *testing.T) {
	t.Parallel()
	c := createFilledMultiCache(time.Second)
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

func TestMultiConcurrentGetSet(t *testing.T) {
	t.Parallel()
	c := createFilledMultiCache(time.Second)
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

func BenchmarkMultiConcurrentGetLRUCache(bb *testing.B) {
	c := createFilledMultiCache(time.Second)
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

func BenchmarkMultiConcurrentSetLRUCache(bb *testing.B) {
	c := createFilledMultiCache(time.Second)
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
func BenchmarkMultiConcurrentSetNXLRUCache(bb *testing.B) {
	c := createFilledMultiCache(time.Second)
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
