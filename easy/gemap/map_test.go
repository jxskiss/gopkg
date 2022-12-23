package gemap

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type dummyStruct struct {
	A string
}

func TestGetTyped(t *testing.T) {
	m1 := Map{"a": dummyStruct{A: "a"}, "b": &dummyStruct{}}
	assert.Equal(t, "a", GetTyped[dummyStruct](m1, "a").A)
	assert.NotNil(t, GetTyped[*dummyStruct](m1, "b"))
}
