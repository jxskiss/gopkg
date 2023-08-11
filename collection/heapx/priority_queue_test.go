package heapx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPriorityQueue(t *testing.T) {
	pq := NewPriorityQueue[int64, *int64]()
	nums := []int64{2, 0, 1, 5, 9, 6, 4, 7, 8, 3}
	for i := range nums {
		num := nums[i]
		pq.Push(&num, num)
	}
	assert.True(t, pq.Len() == 10)
	for i := range nums {
		value, priority, ok := pq.Peek()
		assert.True(t, ok)
		assert.Equal(t, int64(i), *value)
		assert.Equal(t, int64(i), priority)

		value, priority, ok = pq.Pop()
		assert.True(t, ok)
		assert.Equal(t, int64(i), *value)
		assert.Equal(t, int64(i), priority)
	}
}
