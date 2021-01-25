package serialize

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInt64s_Binary32(t *testing.T) {
	slice := Int64List{1, 2, 3, 4, 5}
	buf, _ := slice.MarshalBinary()
	assert.Len(t, buf, 1+4*len(slice))
	assert.Equal(t, binMagic32, buf[0])

	var got Int64List
	err := got.UnmarshalBinary(buf)
	assert.Nil(t, err)
	assert.Equal(t, slice, got)
}

func TestInt64s_Binary64(t *testing.T) {
	slice := Int64List{1, 2, 3, 4, 5, 38194344737811443}
	buf, _ := slice.MarshalBinary()
	assert.Len(t, buf, 1+8*len(slice))
	assert.Equal(t, binMagic64, buf[0])

	var got Int64List
	err := got.UnmarshalBinary(buf)
	assert.Nil(t, err)
	assert.Equal(t, slice, got)
}
