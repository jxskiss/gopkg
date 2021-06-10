package obscure

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestObscure_ID(t *testing.T) {
	key := "hello world"
	obs := New([]byte(key))

	var x int64 = 6590172069002560793
	encoded := obs.EncodeID(x)
	t.Log(string(encoded))

	decoded, err := obs.DecodeID(encoded)
	assert.Nil(t, err)
	assert.Equal(t, x, decoded)
}
