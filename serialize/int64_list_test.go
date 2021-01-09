package serialize

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInt64s_Binary32(t *testing.T) {
	slice := Int64List{1, 2, 3, 4, 5}
	buf, _ := slice.MarshalBinary()
	assert.Len(t, buf, len(binMagic32)+4*len(slice))
	assert.Equal(t, binMagic32, buf[:len(binMagic32)])

	var got Int64List
	err := got.UnmarshalBinary(buf)
	assert.Nil(t, err)
	assert.Equal(t, slice, got)
}

func TestInt64s_Binary64(t *testing.T) {
	slice := Int64List{1, 2, 3, 4, 5, 38194344737811443}
	buf, _ := slice.MarshalBinary()
	assert.Len(t, buf, len(binMagic64)+8*len(slice))
	assert.Equal(t, binMagic64, buf[:len(binMagic64)])

	var got Int64List
	err := got.UnmarshalBinary(buf)
	assert.Nil(t, err)
	assert.Equal(t, slice, got)
}
