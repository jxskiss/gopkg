package reflectx

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type testtype1 struct {
	A string
	B int64
}

type testtype2 struct {
	A string
	B int64
}

type testtype3 struct {
	A string
	B int8
}

func TestIsEqualType(t *testing.T) {
	equal, msg := IsIdenticalType(&testtype1{}, &testtype2{})
	t.Log(msg)
	assert.True(t, equal)

	equal, msg = IsIdenticalType(&testtype1{}, &testtype3{})
	t.Log(msg)
	assert.False(t, equal)

	equal, msg = IsIdenticalType(&testtype2{}, &testtype3{})
	t.Log(msg)
	assert.False(t, equal)
}
