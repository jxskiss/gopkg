package ezmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

type dummyStruct struct {
	A string
}

func TestGetTyped(t *testing.T) {
	m1 := Map{"a": dummyStruct{A: "a"}, "b": &dummyStruct{}}
	assert.Equal(t, "a", GetTyped[dummyStruct](m1, "a").A)
	assert.NotNil(t, GetTyped[*dummyStruct](m1, "b"))
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
