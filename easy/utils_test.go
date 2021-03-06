package easy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetDefault(t *testing.T) {
	intValues := []interface{}{int(1), int32(1), uint16(1), uint64(1)}
	for _, value := range intValues {
		var tmp int16
		SetDefault(&tmp, value)
		assert.Equal(t, int16(1), tmp)
	}

	var ptr *testObject
	var tmp = &testObject{A: 1, B: "b"}
	SetDefault(&ptr, tmp)
	assert.Equal(t, testObject{A: 1, B: "b"}, *ptr)
	assert.Equal(t, tmp, ptr)
}

func TestSetDefault_ShouldPanic(t *testing.T) {
	var ptr *testObject
	var tmp = &testObject{A: 1, B: "b"}

	err := Safe(func() {
		SetDefault(ptr, tmp)
	})()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "SetDefault")
	assert.Contains(t, err.Error(), "must be a non-nil pointer")
}
