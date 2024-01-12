package listx

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStack(t *testing.T) {
	stack := NewStack[int]()
	for i := 0; i < 10; i++ {
		stack.Push(i)
	}
	assert.Equal(t, 10, stack.Len())

	got := make([]int, 0)
	for i := 0; i < 10; i++ {
		x, ok := stack.Pop()
		got = append(got, x)
		assert.True(t, ok)
		if i > 0 {
			assert.Equal(t, got[i-1]-1, got[i])
		}
	}

	_, ok := stack.Peek()
	assert.False(t, ok)
}
