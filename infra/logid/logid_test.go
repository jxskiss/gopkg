package logid

import (
	"runtime"
	"sort"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jxskiss/gopkg/v2/perf/fastrand"
)

func TestDefault(t *testing.T) {
	id1 := Gen()
	assert.Len(t, id1, v1Length)

	SetDefault(NewV2Gen(nil))
	defer SetDefault(NewV1Gen())
	id2 := Gen()
	assert.Len(t, id2, v2IPv4Length)
}

func Test_encodeBase32(t *testing.T) {
	for i := 0; i < 1000; i++ {
		var buf = make([]byte, 10)
		x := fastrand.Uint64()
		encodeBase32(buf, x&mask50bits)
		got, err := decodeBase32(string(buf))
		assert.Nil(t, err)
		assert.Equal(t, x&mask50bits, got)
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
	assert.True(t, dupCount < 2)
}
