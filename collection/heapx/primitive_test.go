package heapx

import (
	"container/heap"
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMinImpl(t *testing.T) {
	var nums MinImpl[int]
	for i := 0; i < 100; i++ {
		x := rand.Intn(80)
		heap.Push(&nums, x)
	}
	var got []int
	for i := 0; i < 100; i++ {
		got = append(got, heap.Pop(&nums).(int))
		if i > 0 {
			assert.GreaterOrEqual(t, got[i], got[i-1])
		}
	}
}

func TestMaxImpl(t *testing.T) {
	var strs MaxImpl[string]
	for i := 0; i < 100; i++ {
		x := fmt.Sprintf("%02d", rand.Intn(80))
		heap.Push(&strs, x)
	}
	var got []string
	for i := 0; i < 100; i++ {
		got = append(got, heap.Pop(&strs).(string))
		if i > 0 {
			assert.GreaterOrEqual(t, got[i-1], got[i])
		}
	}
}
