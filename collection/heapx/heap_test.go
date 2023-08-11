package heapx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func testCmpImpl[T Ordered](a, b T) int {
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}

type cmpIntAsc int

func (x cmpIntAsc) Compare(other cmpIntAsc) int {
	return testCmpImpl(x, other)
}

type cmpIntDesc int

func (x cmpIntDesc) Compare(other cmpIntDesc) int {
	return testCmpImpl(other, x)
}

func TestHeap(t *testing.T) {
	nums := []int{2, 0, 1, 5, 9, 6, 4, 7, 8, 3}

	t.Run("min heap", func(t *testing.T) {
		h1 := NewHeap[cmpIntAsc]()
		assert.True(t, h1.Len() == 0)
		for i := range nums {
			h1.Push(cmpIntAsc(nums[i]))
		}
		for i := range nums {
			x, ok := h1.Pop()
			assert.True(t, ok)
			assert.Equal(t, i, int(x))
		}
		x, ok := h1.Peek()
		assert.False(t, ok)
		assert.Equal(t, 0, int(x))
	})

	t.Run("max heap", func(t *testing.T) {
		h2 := NewHeap[cmpIntDesc]()
		assert.True(t, h2.Len() == 0)
		for i := range nums {
			h2.Push(cmpIntDesc(nums[i]))
		}
		for i := 9; i >= 0; i-- {
			x, ok := h2.Pop()
			assert.True(t, ok)
			assert.Equal(t, i, int(x))
		}
		x, ok := h2.Peek()
		assert.False(t, ok)
		assert.Equal(t, 0, int(x))
	})
}
