package listx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	que := NewQueue[int]()
	for i := 0; i < 10; i++ {
		que.Enqueue(i)
	}
	assert.Equal(t, 10, que.Len())

	got := make([]int, 0)
	for i := 0; i < 10; i++ {
		x, ok := que.Dequeue()
		got = append(got, x)
		assert.True(t, ok)
		if i > 0 {
			assert.Equal(t, got[i-1]+1, got[i])
		}
	}

	_, ok := que.Peek()
	assert.False(t, ok)
}
