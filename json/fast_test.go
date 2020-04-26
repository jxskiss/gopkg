package json

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math"
	"net"
	"reflect"
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
	"nil_value3": testStrSlice(nil),
	"nil_value4": ginH(nil),
	"nil_value5": testSSMap(nil),

	"empty_value1": []int64{},
	"empty_value2": testStrSlice{},
	"empty_value3": ginH{},
	"empty_value4": testSSMap{},

	"typ_slice4": testI32Slice{1, 2, 3},
	"typ_slice5": testStrSlice{"a", "b", "c"},
	"typ_gin_h":  ginH{"a": 1, "b": 2.34, "c": "foobar"},
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
}

func TestMarshalStringMap(t *testing.T) {
	strMap := testStringMap

	_, err := json.Marshal(strMap)
	require.Nil(t, err)

	buf2, err := Marshal(strMap)
	require.Nil(t, err)
	var got2 map[string]string
	err = json.Unmarshal(buf2, &got2)
	require.Nil(t, err)
	assert.Equal(t, strMap, got2)

	buf3, err := Marshal(strMap)
	require.Nil(t, err)
	var got3 map[string]string
	err = json.Unmarshal(buf3, &got3)
	require.Nil(t, err)
	assert.Equal(t, strMap, got3)
}

func TestMarshalStringInterfaceMap(t *testing.T) {
	strMap := testStringInterfaceMap

	buf1, err := json.Marshal(strMap)
	require.Nil(t, err)

	buf2, err := Marshal(strMap)
	require.Nil(t, err)

	var got1 map[string]interface{}
	err = json.Unmarshal(buf1, &got1)
	require.Nil(t, err)
	var got2 map[string]interface{}
	err = json.Unmarshal(buf2, &got2)
	require.Nil(t, err)
	assert.Equal(t, got1, got2)
}

type jsonMarshalerStruct struct {
}

func (p *jsonMarshalerStruct) MarshalJSON() ([]byte, error) {
	return []byte(`"foo"`), nil
}

type textMarshalerStruct struct {
}

func (p *textMarshalerStruct) MarshalText() ([]byte, error) {
	return []byte("bar"), nil
}

func TestNilPointer(t *testing.T) {
	for _, test := range []interface{}{
		(*jsonMarshalerStruct)(nil),
		(*textMarshalerStruct)(nil),
		jsonMarshalerStruct{},
		textMarshalerStruct{},
	} {
		want, err := json.Marshal(test)
		assert.Nil(t, err)
		got, err := Marshal(test)
		assert.Nil(t, err)
		assert.Equal(t, want, got)
	}
}

type testI32Slice []int32

type testStrSlice []string

type ginH map[string]interface{}

type logrusFields map[string]interface{}

type testSSMap map[string]string

func TestTypeAssertion(t *testing.T) {
	strMap := testStringInterfaceMap

	var x1 testI32Slice
	assert.True(t, isIntSlice(reflect.TypeOf(x1)))
	assert.True(t, isIntSlice(reflect.TypeOf(strMap["int_slice1"])))
	assert.True(t, isIntSlice(reflect.TypeOf(strMap["int_slice2"])))
	assert.True(t, isIntSlice(reflect.TypeOf(strMap["int_slice3"])))

	var x2 testStrSlice
	assert.True(t, isStringSlice(reflect.TypeOf(x2)))
	assert.True(t, isStringSlice(reflect.TypeOf(strMap["typ_slice5"])))

	var x3 ginH
	var x4 logrusFields
	assert.True(t, isStringInterfaceMap(reflect.TypeOf(x3)))
	assert.True(t, isStringInterfaceMap(reflect.TypeOf(x4)))
	assert.True(t, isStringInterfaceMap(reflect.TypeOf(strMap["typ_gin_h"])))

	var x5 testSSMap
	assert.True(t, isStringMap(reflect.TypeOf(x5)))
	assert.True(t, isStringMap(reflect.TypeOf(strMap["typ_ss_map"])))
}

func TestMarshalIntSlice(t *testing.T) {
	var x1 = testI32Slice{1, 2, 3}
	got1, _ := Marshal(x1)
	assert.Equal(t, "[1,2,3]", string(got1))
}

func TestMarshalStringSlice(t *testing.T) {
	var x1 = testStrSlice{"a", "b", `"c`}
	got1, _ := Marshal(x1)
	assert.Equal(t, `["a","b","\"c"]`, string(got1))
}

func TestMarshalNilValues(t *testing.T) {
	got, _ := Marshal(nil)
	assert.Equal(t, "null", string(got))

	got, _ = Marshal((testI32Slice)(nil))
	assert.Equal(t, "null", string(got))
	got, _ = Marshal(testStrSlice(nil))
	assert.Equal(t, "null", string(got))
	got, _ = Marshal([]int64{})
	assert.Equal(t, "[]", string(got))

	got, _ = Marshal((ginH)(nil))
	assert.Equal(t, "null", string(got))
	got, _ = Marshal(ginH{})
	assert.Equal(t, "{}", string(got))

	got, _ = Marshal((testSSMap)(nil))
	assert.Equal(t, "null", string(got))
	got, _ = Marshal(testSSMap{})
	assert.Equal(t, "{}", string(got))
}

func TestUnmarshalStringMap(t *testing.T) {
	strMap := testStringMap
	stdstr, err := json.Marshal(strMap)
	require.Nil(t, err)

	var got1 map[string]string
	err = unmarshalStringMap(stdstr, &got1)
	assert.Nil(t, err)
	assert.Equal(t, strMap, got1)

	var got2 map[string]string
	err = Unmarshal(stdstr, &got2)
	assert.Nil(t, err)
	assert.Equal(t, strMap, got2)

	var got3 testSSMap
	assert.True(t, isStringMapPtr(reflect.TypeOf(&got3)))
	err = Unmarshal(stdstr, &got3)
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

func TestMarshaler(t *testing.T) {
	want := `"a:1 b:2 c:3 "`
	var x1 = customJSONMarshaler{
		"a": "1",
		"b": "2",
		"c": "3",
	}
	got1, err := x1.MarshalJSON()
	assert.Nil(t, err)
	assert.Equal(t, want, string(got1))
	got2, err := Marshal(x1)
	assert.Nil(t, err)
	assert.Equal(t, want, string(got2))

	var x2 = customTextMarshaler{
		"a": "1",
		"b": int16(2),
		"c": float64(3),
	}
	got3, err := x2.MarshalText()
	assert.Nil(t, err)
	assert.Equal(t, want[1:len(want)-1], string(got3))
	got4, err := Marshal(x2)
	assert.Nil(t, err)
	assert.Equal(t, want, string(got4))
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
