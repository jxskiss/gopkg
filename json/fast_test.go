package json

import (
	"encoding/json"
	"fmt"
	"github.com/jxskiss/gopkg/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math"
	"net"
	"sort"
	"testing"
)

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

type testSSMap map[string]string

func TestMarshalStringMap(t *testing.T) {
	strMap := testStringMap

	_, err := json.Marshal(strMap)
	require.Nil(t, err)

	buf2, err := MarshalStringMapUnordered(strMap)
	require.Nil(t, err)
	var got2 map[string]string
	err = json.Unmarshal(buf2, &got2)
	require.Nil(t, err)
	assert.Equal(t, strMap, got2)

	buf3, err := MarshalStringMapUnordered(strMap)
	require.Nil(t, err)
	var got3 map[string]string
	err = json.Unmarshal(buf3, &got3)
	require.Nil(t, err)
	assert.Equal(t, strMap, got3)
}

func TestUnmarshalStringMap(t *testing.T) {
	strMap := testStringMap
	stdstr, err := json.Marshal(strMap)
	require.Nil(t, err)

	var got1 map[string]string
	err = UnmarshalStringMap(stdstr, &got1)
	assert.Nil(t, err)
	assert.Equal(t, strMap, got1)

	var got2 map[string]string
	err = Unmarshal(stdstr, &got2)
	assert.Nil(t, err)
	assert.Equal(t, strMap, got2)

	var got3 testSSMap
	err = UnmarshalStringMap(stdstr, (*map[string]string)(&got3))
	assert.Nil(t, err)
	assert.Equal(t, testSSMap(strMap), got3)
}

func TestUnmarshalNull(t *testing.T) {
	var got1 map[string]string
	err := Unmarshal(nullJSON, &got1)
	assert.Nil(t, err)
	assert.Nil(t, got1)

	var got2 testSSMap
	err = Unmarshal(nullJSON, &got2)
	assert.Nil(t, err)
	assert.Nil(t, got2)
}

type customJSONMarshaler map[string]string

func (p customJSONMarshaler) MarshalJSON() ([]byte, error) {
	keys := getSortedStringKeys(p)
	buf := make([]byte, 0)
	buf = append(buf, '"')
	for _, k := range keys {
		tmp := fmt.Sprintf("%v:%v ", k, p[k])
		buf = append(buf, tmp...)
	}
	buf = append(buf, '"')
	return buf, nil
}

type customTextMarshaler map[string]interface{}

func (p customTextMarshaler) MarshalText() ([]byte, error) {
	keys := getSortedStringKeys(p)
	buf := make([]byte, 0)
	for _, k := range keys {
		tmp := fmt.Sprintf("%v:%v ", k, p[k])
		buf = append(buf, tmp...)
	}
	return buf, nil
}

func getSortedStringKeys(m interface{}) []string {
	keys := make([]string, 0)
	switch m := m.(type) {
	case customJSONMarshaler:
		for k := range m {
			keys = append(keys, k)
		}
	case customTextMarshaler:
		for k := range m {
			keys = append(keys, k)
		}
	default:
		panic("not implemented")
	}
	sort.Strings(keys)
	return keys
}
