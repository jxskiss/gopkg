package ezmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeMap(t *testing.T) {
	sm := NewSafeMap()

	func() {
		sm.Set("var1", "value1")
		sm.Set("var2", 1234)
	}()

	assert.Equal(t, "value1", sm.MustGet("var1"))
	assert.Equal(t, 1234, sm.MustGet("var2"))
}
