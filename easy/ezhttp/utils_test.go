package ezhttp

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testObject struct {
	A int    `xml:"a" json:"a"`
	B string `xml:"b" json:"b"`
}

func TestDecodeJSON(t *testing.T) {
	var data testObject
	r := bytes.NewBufferString(`{"a":1,"b":"2"}`)
	err := DecodeJSON(r, &data)
	require.Nil(t, err)

	want := testObject{A: 1, B: "2"}
	assert.Equal(t, want, data)
}

func TestDecodeXML(t *testing.T) {
	var data testObject
	r := bytes.NewBufferString(`<testObject><a>123</a><b>456</b></testObject>`)
	err := DecodeXML(r, &data)
	require.Nil(t, err)

	want := testObject{A: 123, B: "456"}
	assert.Equal(t, want, data)
}
