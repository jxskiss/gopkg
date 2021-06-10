package obscure

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestObscure_Reproducible(t *testing.T) {
	key := "hello world"
	obs1 := New([]byte(key))
	obs2 := New([]byte(key))
	assert.Equal(t, obs1.Index(), obs2.Index())

	table1 := obs1.Table()
	table2 := obs2.Table()
	for i := 0; i < idxlen; i++ {
		assert.Equal(t, table1[i], table2[i])
	}
}

func TestObscure(t *testing.T) {
	key := "hello world"
	obs := New([]byte(key))

	encoded := obs.EncodeToBytes([]byte(key))
	decoded, err := obs.DecodeBytes(encoded)
	assert.Nil(t, err)
	assert.Equal(t, key, string(decoded))
}
