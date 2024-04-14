package ezmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalCSV(t *testing.T) {
	type AString string
	type AStruct struct {
		A bool
		B AString
	}
	records := []Map{
		{
			"bool":   true,
			"int":    1234,
			"str":    "12345",
			"str_2":  AString("23456"),
			"struct": AStruct{A: true, B: "23456"},
		},
		{
			"int":    4321,
			"str_2":  AString("65432"),
			"bool":   false,
			"struct": AStruct{A: false, B: "65432"},
			"str":    "12345",
		},
	}
	got, err := MarshalCSV(records)
	require.Nil(t, err)

	want := `bool,int,str,str_2,struct
true,1234,12345,23456,"{""A"":true,""B"":""23456""}"
false,4321,12345,65432,"{""A"":false,""B"":""65432""}"
`
	assert.Equal(t, want, string(got))
}
