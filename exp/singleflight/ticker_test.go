package singleflight

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCallbackTicker(t *testing.T) {
	ticker1 := newCallbackTicker(time.Second, func() {})
	assert.IsType(t, &stdTickerImpl{}, ticker1)

	ticker2 := newManySelectTicker(time.Second, func() {})
	assert.IsType(t, &manySelectTickerImpl{}, ticker2)

	time.Sleep(100 * time.Millisecond)
	ticker1.Stop()
	ticker2.Stop()
}

func TestStdTickerImpl(t *testing.T) {
	var count int32
	ticker := newStdTicker(100*time.Millisecond, func() {
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

func TestManySelectTickerImpl(t *testing.T) {
	var count int32
	ticker := newManySelectTicker(100*time.Millisecond, func() {
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
