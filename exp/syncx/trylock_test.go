package syncx

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTryLock_Lock(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()
	m := NewTryLock()
	assert.Nil(t, m.Lock(ctx))
	assert.True(t, m.Lock(ctx) != nil)
	m.Unlock()
	assert.Nil(t, m.Lock(context.TODO()))
}

func TestTryLock_TryLock(t *testing.T) {
	m := NewTryLock()
	assert.Nil(t, m.Lock(context.TODO()))
	assert.False(t, m.TryLock())
	m.Unlock()
	assert.True(t, m.TryLock())
}

var (
	rCount    = 100
	sleepTime = 100 * time.Millisecond
	totalTime = time.Duration(rCount) * sleepTime
)

func TestTryLock(t *testing.T) {
	start := time.Now()
	m := NewTryLock()
	value := 0
	wg := sync.WaitGroup{}
	for i := 0; i < rCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = m.Lock(context.TODO())
			defer m.Unlock()
			assert.Equal(t, value, 0)
			value = 1
			time.Sleep(sleepTime)
			value = 0
		}()
	}
	wg.Wait()
	t.Log(time.Now().Sub(start), totalTime)
	assert.True(t, time.Now().Sub(start) >= totalTime)
}

func BenchmarkTryLock_Lock(b *testing.B) {
	lock := NewTryLock()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = lock.Lock(context.TODO())
			//nolint:staticcheck
			lock.Unlock()
		}
	})
}
