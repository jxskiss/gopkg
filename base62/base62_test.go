package base62

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestEncodeDecode(t *testing.T) {
	buf := make([]byte, rand.Intn(64))
	_, _ = rand.Read(buf)
	encResult := EncodeToString(buf)
	decResult, err := DecodeString(encResult)
	assert.Nil(t, err)
	assert.Equal(t, buf, decResult)
}
