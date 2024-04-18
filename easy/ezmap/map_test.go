package ezmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestMap(t *testing.T) {
	t.Run("zero map", func(t *testing.T) {
		var m Map
		val1, exists1 := m.Get("abc")
		assert.False(t, exists1)
		assert.Nil(t, val1)
		assert.Nil(t, m.GetOr("abc", nil))
		assert.Panics(t, func() { _ = m.MustGet("abc") })

		m.Set("abc", "abc")
		val2, exists2 := m.Get("abc")
		assert.True(t, exists2)
		assert.Equal(t, "abc", val2)
	})

	t.Run("Merge", func(t *testing.T) {
		m1 := Map{"abc": 123}
		var m2 Map
		m2.Merge(m1)
		require.NotNil(t, m2)
		assert.Equal(t, 123, m2.GetOr("abc", 0))
	})
}

func TestYAMLMarshaling(t *testing.T) {
	s := `
servers:
  - ports: [ 80, ]
    server_names:
      - "www.example1.com"
      - "www.example2.com"
    services:
      - "serviceName1"
  - ports: [ 80, 81 ]
    server_names:
      - "api.example1.com"
    services:
      - "serviceName2"
`
	var m Map
	err := yaml.Unmarshal([]byte(s), &m)
	require.Nil(t, err)

	services, ok := m.GetSlice("servers").([]any)
	require.True(t, ok)
	assert.Len(t, services, 2)
	assert.Equal(t, 80, services[0].(map[string]any)["ports"].([]any)[0])
	assert.Equal(t, 81, services[1].(map[string]any)["ports"].([]any)[1])
	assert.Equal(t, "serviceName2", services[1].(map[string]any)["services"].([]any)[0])
}
