package ezmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeMap(t *testing.T) {
	sm := NewSafeMap()

	func() {
		sm.Lock()
		defer sm.Unlock()
		sm.Set("var1", "value1")
		sm.Set("var2", 1234)
	}()

	sm.RLock()
	assert.Equal(t, "value1", sm.MustGet("var1"))
	assert.Equal(t, 1234, sm.MustGet("var2"))
	assert.Equal(t, "value1", GetTyped[string](sm.Map, "var1"))
	assert.Equal(t, 1234, GetTyped[int](sm.Map, "var2"))
}
