package ptr

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCopyAny(t *testing.T) {
	x1 := int64(1)

	got1, ok := CopyAny(x1).(*int64)
	assert.True(t, ok)
	assert.Equal(t, x1, *got1)

	got2, ok := CopyAny(&x1).(*int64)
	assert.True(t, ok)
	assert.Equal(t, x1, *got2)

	got3, ok := CopyAny((*int64)(nil)).(*int64)
	assert.True(t, ok)
	assert.Nil(t, got3)
}
