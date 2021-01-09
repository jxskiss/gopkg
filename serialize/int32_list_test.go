package serialize

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInt32s_Binary(t *testing.T) {
	slice := Int32List{1, 2, 3, 4, 5}
	buf, _ := slice.MarshalBinary()
	assert.Len(t, buf, len(binMagic32)+4*len(slice))
	assert.Equal(t, binMagic32, buf[:len(binMagic32)])

	var got Int32List
	err := got.UnmarshalBinary(buf)
	assert.Nil(t, err)
	assert.Equal(t, slice, got)
}
