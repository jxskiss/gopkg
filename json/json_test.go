package json

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestStringConversion(t *testing.T) {
	want := testStringMap
	str, err := MarshalToString(testStringMap)
	assert.Nil(t, err)
	var got map[string]string
	err = UnmarshalFromString(str, &got)
	assert.Nil(t, err)
	assert.Equal(t, want, got)
}

func TestGet(t *testing.T) {
	data, _ := Marshal(testStringInterfaceMap)

	got1 := Get(data, "ip_address")
	assert.Equal(t, "26.58.193.2", got1.ToString())

	got2 := Get(data, "int_slice3", 2)
	assert.Equal(t, int64(9), got2.ToInt64())

	got3 := Get(data, "not_exists", 0)
	assert.NotNil(t, got3.LastError())
}

func TestGetByDot(t *testing.T) {
	data, _ := Marshal(testStringInterfaceMap)

	got1 := GetByDot(data, "ip_address")
	assert.Equal(t, "26.58.193.2", got1.ToString())

	got2 := GetByDot(data, "some_struct.b_i64")
	assert.Equal(t, int64(456), got2.ToInt64())

	got3 := GetByDot(data, "int_slice3.2")
	assert.Equal(t, 9, got3.ToInt())

	got4 := GetByDot(data, "typ_struct_slice.*.b").GetInterface().([]Any)
	for _, x := range got4 {
		assert.Nil(t, x.LastError())
		assert.Greater(t, x.ToInt(), 0)
	}

	got5 := GetByDot(data, "typ_struct_slice.a")
	assert.NotNil(t, got5.LastError())
}
