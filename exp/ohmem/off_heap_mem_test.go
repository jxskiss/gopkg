package ohmem

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestOffHeapMem(t *testing.T) {
	ohm := NewOffHeapMem([]Class{
		{4096, 512},
		{2048, 512},
		{1024, 512},
		{1024, 128},
		{512, 512},
		{128, 512},
		{128, 64},
	})
	alloc := ohm.Alloc
	free := ohm.Free

	n := int(1e5)
	t0 := time.Now()
	wg := new(sync.WaitGroup)
	wg.Add(runtime.NumCPU())
	for range make([]struct{}, runtime.NumCPU()) {
		go func() {
			for i := 0; i < n; i++ {
				// 分配
				bs := alloc(1 + rand.Intn(4095))
				// 回收
				free(bs)
			}
			wg.Done()
		}()
	}
	wg.Wait()
	fmt.Printf("%v\n", time.Since(t0)/time.Duration(n*runtime.NumCPU()))
}
