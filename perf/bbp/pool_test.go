package bbp

import (
	"math/rand"
	"testing"
	"time"
)

func TestPoolCalibrate(t *testing.T) {
	calls := defaultCalibrateCalls
	for i := 0; i < poolSize*calls; i++ {
		n := 1004
		if i%15 == 0 {
			n = rand.Intn(15234)
		}
		testGetPut(t, n)
	}
}

func TestPoolVariousSizesSerial(t *testing.T) {
	testPoolVariousSizes(t)
}

func TestPoolVariousSizesConcurrent(t *testing.T) {
	concurrency := 5
	ch := make(chan struct{}, 5)
	for i := 0; i < concurrency; i++ {
		go func() {
			testPoolVariousSizes(t)
			ch <- struct{}{}
		}()
	}
	timeout := 10 * time.Second
	for i := 0; i < concurrency; i++ {
		select {
		case <-ch:
		case <-time.After(timeout):
			t.Fatalf("%v timeout", timeout)
		}
	}
}

func testPoolVariousSizes(t *testing.T) {
	for i := 0; i < poolSize-minPoolIdx; i++ {
		n := bufSizeTable[i]

		testGetPut(t, n)
		testGetPut(t, n+1)
		testGetPut(t, n-1)
	}
}

var testPool Pool

func testGetPut(t *testing.T, n int) {
	bb := testPool.Get()
	if len(bb) > 0 {
		t.Fatalf("non-empty byte buffer returned from acquire")
	}
	bb = allocNBytes(bb, n)
	testPool.Put(bb)
}

func allocNBytes(dst []byte, n int) []byte {
	diff := n - cap(dst)
	if diff <= 0 {
		return dst[:n]
	}
	return append(dst, make([]byte, diff)...)
}
