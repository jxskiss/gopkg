package functicker

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestTicker(t *testing.T) {
	var count atomic.Int32
	ticker := New(100*time.Millisecond, func() {
		count.Add(1)
		t.Log("functicker.Ticker tick")
	})
	time.Sleep(time.Second)

	ticker.Reset(200 * time.Millisecond)
	time.Sleep(time.Second)

	ticker.Stop()
	time.Sleep(time.Second)

	if n := count.Load(); !(n > 10 && n <= 15) {
		t.Errorf("got unexpected count %d", n)
	}
}
