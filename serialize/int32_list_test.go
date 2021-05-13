package serialize

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInt32s_Binary(t *testing.T) {
	slice := Int32List{1, 2, 3, 4, 5}
	buf, _ := slice.MarshalBinary()
	assert.Len(t, buf, 1+4*len(slice))
	assert.Equal(t, binMagic32, buf[0])

	var got Int32List
	err := got.UnmarshalBinary(buf)
	assert.Nil(t, err)
	assert.Equal(t, slice, got)
}

func TestInt32s_DiffCompressed(t *testing.T) {
	slice := Int32List{123, 456, 345, 789, 678}
	buf, _ := slice.MarshalDiffCompressed()
	t.Logf("Int32List DiffCompressed, len(buf)= %v", len(buf))
	assert.Equal(t, binDiffCompressed, buf[0])

	var got Int32List
	err := got.UnmarshalBinary(buf)
	assert.Nil(t, err)
	assert.Equal(t, slice, got)
}
