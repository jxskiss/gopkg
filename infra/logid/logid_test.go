package logid

import (
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func BenchmarkGen(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Gen()
	}
}

func TestUniqueness(t *testing.T) {
	var count int32 = -1
	var got = make([]string, 1000)

	var wg sync.WaitGroup
	var n = runtime.GOMAXPROCS(0)
	for j := 0; j < n; j++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				i := atomic.AddInt32(&count, 1)
				if int(i) < len(got) {
					got[i] = Gen()
					continue
				}
				break
			}
		}()
	}
	wg.Wait()

	sort.Strings(got)
	dupCount := 0
	for i := 0; i < len(got)-1; i++ {
		if got[i] == got[i+1] {
			dupCount++
		}
	}
	assert.True(t, dupCount < 3)
}
