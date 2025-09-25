package ezmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeMap(t *testing.T) {
	var sm SafeMap
	func() {
		sm.Set("var1", "value1")
		sm.Set("var2", 1234)
		sm.Set("slice1", []int{1, 2, 3, 4})
		sm.Set("slice2", []any{
			Map{"a": 1},
			map[string]any{"b": 2},
			map[string]string{"c": "3"},
		})
	}()

	assert.Equal(t, "value1", sm.MustGet("var1"))
	assert.Equal(t, "value1", sm.GetString("var1"))
	assert.Equal(t, 1234, sm.MustGet("var2"))
	assert.Equal(t, 1234, sm.GetInt("var2"))
	assert.Equal(t, int32(1234), sm.GetInt32("var2"))
	assert.Equal(t, int64(1234), sm.GetInt64("var2"))
	assert.Equal(t, []any{1, 2, 3, 4}, sm.GetSlice("slice1"))
	assert.Equal(t, 1, sm.GetSliceElem("slice1", 0))
	assert.Equal(t, Map{"a": 1}, sm.GetSliceElemMap("slice2", 0))
	assert.Equal(t, Map{"b": 2}, sm.GetSliceElemMap("slice2", 1))
	assert.Nil(t, sm.GetSliceElemMap("slice2", 2))
}

func TestSafeMap_WithLock(t *testing.T) {
	var sm SafeMap
	sm.WithLock(func(m Map) {
		m.Set("var1", "value1")
	})
	assert.Equal(t, "value1", sm.MustGet("var1"))

	var got1 string
	sm.WithRLock(func(m Map) {
		got1 = m.GetString("var1")
	})
	assert.Equal(t, "value1", got1)
}
