package syncx

import (
	"io"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRWLock(t *testing.T) {
	var wg sync.WaitGroup
	lock := NewRWLock()
	count := 0
	for i := 0; i < 1000; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			for i := 0; i < 10; i++ {
				l := lock.RLock()
				s := strconv.Itoa(count)
				_, _ = io.Discard.Write([]byte(s))
				l.Unlock()
			}
		}()
		go func() {
			defer wg.Done()
			lock.Lock()
			count++
			lock.Unlock()
		}()
	}
	wg.Wait()
	assert.Equal(t, 1000, count)
}
