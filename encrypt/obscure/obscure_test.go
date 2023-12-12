package obscure

import (
	cryptorand "crypto/rand"
	mathrand "math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestObscure_Reproducible(t *testing.T) {
	key := "hello world"
	obs1 := New([]byte(key))
	obs2 := New([]byte(key))
	assert.Equal(t, obs1.Index(), obs2.Index())

	table1 := obs1.Table()
	table2 := obs2.Table()
	for i := 0; i < idxLen; i++ {
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

func TestObscure_ZeroLengthInput(t *testing.T) {
	key := []byte{248, 77, 8, 37, 138, 1, 232, 104, 31}
	obs := New(key)
	data := []byte{}
	encoded := obs.EncodeToString(data)
	decoded, err := obs.DecodeString(encoded)
	assert.Nil(t, err)
	assert.Nil(t, decoded)
}

func TestObscure_RandomBytes(t *testing.T) {
	for i := 0; i < 1000; i++ {
		key := make([]byte, mathrand.Intn(100))
		_, _ = cryptorand.Read(key)
		obs := New(key)

		data := make([]byte, 1000)
		for j := 0; j < 1000; j++ {
			data = data[:mathrand.Intn(1000)]
			_, _ = cryptorand.Read(data)
			encoded := obs.EncodeToString(data)
			decoded, err := obs.DecodeString(encoded)
			assert.Nil(t, err)
			if len(data) == 0 {
				assert.Nil(t, decoded)
			} else {
				assert.Equal(t, data, decoded)
			}
		}
	}
}
