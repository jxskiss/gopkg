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

func TestUnmarshalCSV(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		data := `bool,int,str,str_2,struct
true,1234,12345,23456,"{""A"":true,""B"":""23456""}"
false,4321,12345,65432,"{""A"":false,""B"":""65432""}"
`
		got, err := UnmarshalCVS([]byte(data))
		require.Nil(t, err)

		want := []Map{
			{
				"bool":   "true",
				"int":    "1234",
				"str":    "12345",
				"str_2":  "23456",
				"struct": `{"A":true,"B":"23456"}`,
			},
			{
				"int":    "4321",
				"str_2":  "65432",
				"bool":   "false",
				"struct": `{"A":false,"B":"65432"}`,
				"str":    "12345",
			},
		}
		assert.Equal(t, want, got)
	})

	t.Run("duplicate header", func(t *testing.T) {
		data := `bool,int,str,int
true,123,"abc",456
`
		_, err := UnmarshalCVS([]byte(data))
		require.NotNil(t, err)
		assert.Contains(t, err.Error(), "duplicate header: int")
	})
}
