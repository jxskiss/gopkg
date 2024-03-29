package heapx

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHeap(t *testing.T) {
	nums := make([]int, 3333)
	for i := range nums {
		nums[i] = i
	}
	rand.Shuffle(len(nums), func(i, j int) {
		nums[i], nums[j] = nums[j], nums[i]
	})

	t.Run("min heap", func(t *testing.T) {
		h1 := NewHeap[int](func(lhs, rhs int) bool {
			return lhs < rhs
		})
		assert.True(t, h1.Len() == 0)
		for i := range nums {
			h1.Push(nums[i])
		}
		for i := range nums {
			x, ok := h1.Pop()
			assert.True(t, ok)
			assert.Equal(t, i, x)
		}
		x, ok := h1.Peek()
		assert.False(t, ok)
		assert.Equal(t, 0, x)
		assert.Equal(t, 1, len(h1.items.ss))
		assert.Equal(t, bktSize, h1.items.cap)
		assert.Equal(t, 0, h1.items.len)
	})

	t.Run("max heap", func(t *testing.T) {
		h2 := NewHeap[int](func(lhs, rhs int) bool {
			return rhs < lhs
		})
		assert.True(t, h2.Len() == 0)
		for i := range nums {
			h2.Push(nums[i])
		}
		for i := len(nums) - 1; i >= 0; i-- {
			x, ok := h2.Pop()
			assert.True(t, ok)
			assert.Equal(t, i, x)
		}
		x, ok := h2.Peek()
		assert.False(t, ok)
		assert.Equal(t, 0, x)
		assert.Equal(t, 1, len(h2.items.ss))
		assert.Equal(t, bktSize, h2.items.cap)
		assert.Equal(t, 0, h2.items.len)
	})

	t.Run("shrink", func(t *testing.T) {
		h3 := NewHeap[int](func(lhs, rhs int) bool {
			return lhs < rhs
		})
		assert.Equal(t, 0, h3.Len())
		assert.Equal(t, 0, h3.items.cap)
		assert.Equal(t, 0, len(h3.items.ss))

		for i := 0; i < bktSize; i++ {
			h3.Push(i)
		}
		assert.Equal(t, bktSize, h3.Len())
		assert.Equal(t, bktSize, h3.items.cap)
		assert.Equal(t, 1, len(h3.items.ss))

		for i := 0; i < bktSize; i++ {
			h3.Push(i)
		}
		assert.Equal(t, 2*bktSize, h3.Len())
		assert.Equal(t, 2*bktSize, h3.items.cap)
		assert.Equal(t, 2, len(h3.items.ss))

		for i := 0; i < 10; i++ {
			h3.Push(i)
		}
		assert.Equal(t, 2*bktSize+10, h3.Len())
		assert.Equal(t, 3*bktSize, h3.items.cap)
		assert.Equal(t, 3, len(h3.items.ss))

		for i := 0; i < 10; i++ {
			h3.Pop()
		}
		assert.Equal(t, 2*bktSize, h3.Len())
		assert.Equal(t, 3*bktSize, h3.items.cap)
		assert.Equal(t, 3, len(h3.items.ss))

		for i := 0; i < shrinkThreshold-1; i++ {
			h3.Pop()
		}
		assert.Equal(t, 2*bktSize-(shrinkThreshold-1), h3.Len())
		assert.Equal(t, 3*bktSize, h3.items.cap)
		assert.Equal(t, 3, len(h3.items.ss))

		h3.Pop()
		assert.Equal(t, 2*bktSize-shrinkThreshold, h3.Len())
		assert.Equal(t, 2*bktSize, h3.items.cap)
		assert.Equal(t, 2, len(h3.items.ss))

		for i := 0; i < (bktSize - shrinkThreshold); i++ {
			h3.Pop()
		}
		assert.Equal(t, bktSize, h3.Len())
		assert.Equal(t, 2*bktSize, h3.items.cap)
		assert.Equal(t, 2, len(h3.items.ss))

		for i := 0; i < shrinkThreshold-1; i++ {
			h3.Pop()
		}
		assert.Equal(t, bktSize-(shrinkThreshold-1), h3.Len())
		assert.Equal(t, 2*bktSize, h3.items.cap)
		assert.Equal(t, 2, len(h3.items.ss))

		h3.Pop()
		assert.Equal(t, bktSize-shrinkThreshold, h3.Len())
		assert.Equal(t, bktSize, h3.items.cap)
		assert.Equal(t, 1, len(h3.items.ss))

		for i := 0; i < bktSize-shrinkThreshold; i++ {
			h3.Pop()
		}
		assert.Equal(t, 0, h3.Len())
		assert.Equal(t, bktSize, h3.items.cap)
		assert.Equal(t, 1, len(h3.items.ss))
	})
}
