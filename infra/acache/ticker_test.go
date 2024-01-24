package acache

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCallbackTicker(t *testing.T) {
	ticker1 := newCallbackTicker(time.Second, func(_ time.Time, _ bool) {})
	assert.IsType(t, &tickerImpl{}, ticker1)

	time.Sleep(100 * time.Millisecond)
	ticker1.Stop()
}

func TestTickerImpl(t *testing.T) {
	var count int32
	ticker := newTicker(100*time.Millisecond, func(_ time.Time, _ bool) {
		atomic.AddInt32(&count, 1)
	})

	time.Sleep(1100 * time.Millisecond)
	ticker.Stop()
	n1 := atomic.LoadInt32(&count)
	assert.True(t, n1 >= 9 && n1 <= 11)

	time.Sleep(300 * time.Millisecond)
	n2 := atomic.LoadInt32(&count)
	assert.True(t, n2 >= 9 && n2 <= 11)
}
