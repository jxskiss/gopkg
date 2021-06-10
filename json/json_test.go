package json

import (
	"bytes"
	stdjson "encoding/json"
	"github.com/jxskiss/gopkg/ptr"
	"github.com/stretchr/testify/assert"
	"math"
	"net"
	"testing"
)

type testSSMap map[string]string

var testStringMap = map[string]string{
	"id":         "1234567",
	"first_name": "Jeanette",
	"last_name":  "Penddreth",
	"email":      "jpenddreth0@census.gov",
	"gender":     "Female",
	"ip_address": "26.58.193.2",
	"html_tag":   "<html></html>",
	"chinese":    "北京欢迎你！",
	`a:\b":"\"c`: `d\"e:f`,
}

var testStringInterfaceMap = map[string]interface{}{
	"id":         12345,
	"id2":        uint(12345),
	"first_name": "Jeanette",
	"last_name":  "Penddreth",
	"email":      "jpenddreth0@census.gov",
	"gender":     "Female",
	"ip_address": net.ParseIP("26.58.193.2"),
	"html_tag":   "<html></html>",
	"chinese":    []string{"北京欢迎你！ \b", "Bejing welcome you!\t\b\n\b"},
	`a:\b":"\"c`: `d\"e:f`,
	"some_struct": struct {
		A int32  `json:"a_i32,omitempty"`
		B int64  `json:"b_i64,omitempty"`
		C string `json:"c_str,omitempty"`
	}{
		B: 456,
		C: "foobar",
	},
	"some_byte_slice": []byte("北京欢迎你！Bejing welcome hou!"),

	"int_slice1":  []int{1, 2, 3},
	"int_slice2":  []int32{4, 5, 6},
	"int_slice3":  []int64{7, 8, 9},
	"uint_slice4": []uint8{1, 127, 253},
	"uint_slice5": []uint64{math.MaxUint64},

	"nil_value1": []int(nil),
	"nil_value2": []uint16(nil),
	"nil_value5": testSSMap(nil),

	"empty_value1": []int64{},
	"empty_value4": testSSMap{},

	"typ_ss_map": testSSMap{"a": "1", "b": "2", "c": "3"},

	"typ_struct_slice": []struct {
		A string `json:"a"`
		B int    `json:"b"`
		C bool   `json:"c"`
	}{
		{"1", 1, true},
		{"2", 2, false},
		{"3", 3, true},
		{"4", 4, false},
	},

	"bool1": true,
	"bool2": ptr.Bool(false),
	"bool3": (*bool)(nil),

	"integer1": int32(1234),
	"integer2": ptr.Int32(1234),
	"integer3": (*int16)(nil),

	"slice_fast_nil0": [][]int32(nil),
	"slice_fast_nil1": [][]string(nil),
	"slice_fast_nil2": []map[string]string(nil),
	"slice_fast_nil3": []map[string]interface{}(nil),

	"slice_fast_typ0": [][]int32{
		{1, 2, 3},
		{4, 5, 6},
	},
	"slice_fast_typ1": [][]string{
		{"a", "b", "c"},
		{"foo", "bar"},
	},
	"slice_fast_typ2": []map[string]string{
		{"a": "1"},
		{"b": "2"},
	},
	"slice_fast_typ3": []map[string]interface{}{
		{"a": "1", "b": 1},
		{"c": "2", "d": 2},
		{"e": int64(3), "d": ptr.Uint64(3)},
		{"f": true, "g": ptr.Bool(false)},
	},
	"slice_fast_typ4": []map[string]interface{}{},

	"slice_fast_typ5": []bool{true, false},
	"slice_fast_typ6": []*bool{ptr.Bool(true), ptr.Bool(false), nil},
	"slice_fast_typ7": []int16{7, 8},
	"slice_fast_typ8": []*int16{ptr.Int16(8), ptr.Int16(9), nil},
}

func TestStringConversion(t *testing.T) {
	want := testStringMap
	str, err := MarshalToString(testStringMap)
	assert.Nil(t, err)

	var got map[string]string
	err = UnmarshalFromString(str, &got)
	assert.Nil(t, err)
	assert.Equal(t, want, got)
}

func TestCompatibility(t *testing.T) {
	stdOutput, err := stdjson.Marshal(testStringInterfaceMap)
	assert.Nil(t, err)

	thisOutput, err := Marshal(testStringInterfaceMap)
	assert.Nil(t, err)
	assert.Equal(t, stdOutput, thisOutput)

	var got1 map[string]interface{}
	var got2 map[string]interface{}
	err = stdjson.Unmarshal(stdOutput, &got1)
	assert.Nil(t, err)
	err = Unmarshal(thisOutput, &got2)
	assert.Nil(t, err)
	assert.Equal(t, got1, got2)
}

func TestCompatibility_NoHTMLEscape_Indent(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	stdenc := stdjson.NewEncoder(buf)
	stdenc.SetEscapeHTML(false)
	stdenc.SetIndent("  ", "  ")
	err := stdenc.Encode(testStringInterfaceMap)
	assert.Nil(t, err)
	stdOutput := buf.Bytes()

	thisOutput, err := MarshalNoHTMLEscape(testStringInterfaceMap, "  ", "  ")
	assert.Nil(t, err)

	assert.Equal(t, stdOutput, thisOutput)
}
