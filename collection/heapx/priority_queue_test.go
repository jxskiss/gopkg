package heapx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPriorityQueue(t *testing.T) {
	nums := []int64{2, 0, 1, 5, 9, 6, 4, 7, 8, 3}

	t.Run("min queue", func(t *testing.T) {
		pq := NewMinPriorityQueue[int64, *int64]()
		for i := range nums {
			num := nums[i]
			pq.Push(num, &num)
		}
		assert.True(t, pq.Len() == 10)
		for i := range nums {
			priority, value, ok := pq.Peek()
			assert.True(t, ok)
			assert.Equal(t, int64(i), *value)
			assert.Equal(t, int64(i), priority)

			priority, value, ok = pq.Pop()
			assert.True(t, ok)
			assert.Equal(t, int64(i), *value)
			assert.Equal(t, int64(i), priority)
		}
	})

	t.Run("max queue", func(t *testing.T) {
		pq := NewMaxPriorityQueue[int64, *int64]()
		for i := range nums {
			num := nums[i]
			pq.Push(num, &num)
		}
		assert.True(t, pq.Len() == 10)
		for i := range nums {
			priority, value, ok := pq.Peek()
			assert.True(t, ok)
			assert.Equal(t, int64(9-i), *value)
			assert.Equal(t, int64(9-i), priority)

			priority, value, ok = pq.Pop()
			assert.True(t, ok)
			assert.Equal(t, int64(9-i), *value)
			assert.Equal(t, int64(9-i), priority)
		}
	})
}
