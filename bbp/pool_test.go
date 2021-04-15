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
	ch := make(chan struct{})
	for i := 0; i < concurrency; i++ {
		go func() {
			testPoolVariousSizes(t)
			ch <- struct{}{}
		}()
	}
	for i := 0; i < concurrency; i++ {
		select {
		case <-ch:
		case <-time.After(time.Second):
			t.Fatalf("timeout")
		}
	}
}

func testPoolVariousSizes(t *testing.T) {
	for i := 0; i < minBufSize+1; i++ {
		n := 1 << i

		testGetPut(t, n)
		testGetPut(t, n+1)
		testGetPut(t, n-1)

		for j := 0; j < 10; j++ {
			testGetPut(t, j+n)
		}
	}
}

var testPool Pool

func testGetPut(t *testing.T, n int) {
	bb := testPool.Get()
	if len(bb.B) > 0 {
		t.Fatalf("non-empty byte buffer returned from acquire")
	}
	bb.B = allocNBytes(bb.B, n)
	testPool.Put(bb)
}

func allocNBytes(dst []byte, n int) []byte {
	diff := n - cap(dst)
	if diff <= 0 {
		return dst[:n]
	}
	return append(dst, make([]byte, diff)...)
}
