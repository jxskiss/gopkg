package ezmap

import (
	"testing"

	"github.com/mitchellh/mapstructure"
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

	t.Run("GetSlice", func(t *testing.T) {
		var m Map
		m.Set("k1", []int{1, 2, 3})
		got := m.GetSlice("k1")
		assert.Len(t, got, 3)
	})

	t.Run("GetSliceElem", func(t *testing.T) {
		var m Map
		m.Set("k1", []int{1, 2, 3})
		assert.Equal(t, 1, m.GetSliceElem("k1", 0))
		assert.Equal(t, 2, m.GetSliceElem("k1", 1))
		assert.Equal(t, 3, m.GetSliceElem("k1", 2))
		assert.Nil(t, m.GetSliceElem("k1", 3))
	})

	t.Run("Merge", func(t *testing.T) {
		m1 := Map{"abc": 123}
		var m2 Map
		m2.Merge(m1)
		require.NotNil(t, m2)
		assert.Equal(t, 123, m2.GetOr("abc", 0))
	})

	t.Run("DecodeToStruct", func(t *testing.T) {
		var m Map
		m.Set("abc", 123)
		m.Set("def", map[string]any{"abc": 456, "Def": "def"})
		var s1 struct {
			Abc int `mapstructure:"abc"`
			Def struct {
				Abc int `mapstructure:"abc"`
				Def string
			} `mapstructure:"def"`
		}

		err1 := m.DecodeToStruct(&s1, nil)
		require.Nil(t, err1)
		assert.Equal(t, 123, s1.Abc)
		assert.Equal(t, 456, s1.Def.Abc)
		assert.Equal(t, "def", s1.Def.Def)

		var s2 struct{ Def string }
		err2 := m.GetMap("def").
			DecodeToStruct(&s2, &mapstructure.DecoderConfig{})
		require.Nil(t, err2)
		assert.Equal(t, "def", s2.Def)
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

	services := m.GetSlice("servers")
	assert.Len(t, services, 2)
	assert.Equal(t, 80, services[0].(map[string]any)["ports"].([]any)[0])
	assert.Equal(t, 81, services[1].(map[string]any)["ports"].([]any)[1])
	assert.Equal(t, "serviceName2", services[1].(map[string]any)["services"].([]any)[0])
}
