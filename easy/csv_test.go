package easy

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jxskiss/gopkg/v2/easy/ezmap"
)

func TestMarshalCSV(t *testing.T) {
	type AString string
	type AStruct struct {
		A bool
		B AString
	}
	records := []ezmap.Map{
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

		want := []ezmap.Map{
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

func TestUnmarshalCSVWithSeparator(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		data := `Name	ID	UID
name1	515211140	17592701255556
name2	502359508	17592688403924
name3	35184904144031	35184904144031
`
		got, err := UnmarshalCSVWithSeparator([]byte(data), '\t')
		require.Nil(t, err)

		want := []ezmap.Map{
			{
				"Name": "name1",
				"ID":   "515211140",
				"UID":  "17592701255556",
			},
			{
				"Name": "name2",
				"ID":   "502359508",
				"UID":  "17592688403924",
			},
			{
				"Name": "name3",
				"ID":   "35184904144031",
				"UID":  "35184904144031",
			},
		}
		assert.Equal(t, want, got)
	})
}
