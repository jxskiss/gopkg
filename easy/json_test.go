package easy

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestJSONMarshalMapInterfaceInterface(t *testing.T) {
	m := make(map[any]any)
	m[1] = "1"
	m["2"] = 2
	got := JSON(m)
	want := `{"1":"1","2":2}`
	assert.Equal(t, want, got)
}

func TestJSONDisableEscapeHTML(t *testing.T) {
	m := map[string]string{
		"html": "<html></html>",
	}

	stdRet, err := json.Marshal(m)
	assert.Nil(t, err)
	assert.Equal(t, `{"html":"\u003chtml\u003e\u003c/html\u003e"}`, string(stdRet))

	got := JSON(m)
	assert.Equal(t, `{"html":"<html></html>"}`, got)
}

func TestLazyJSON(t *testing.T) {
	var x = &testObject{A: 123, B: "abc"}
	got1 := JSON(x)
	got2 := fmt.Sprintf("%v", LazyJSON(x))
	assert.Equal(t, got1, got2)
}

var prettyTestWant = strings.TrimSpace(`
{
    "1": 123,
    "b": "<html>"
}`)

func TestPretty(t *testing.T) {
	test := map[string]any{
		"1": 123,
		"b": "<html>",
	}
	jsonString := JSON(test)
	assert.Equal(t, `{"1":123,"b":"<html>"}`, jsonString)

	got1 := Pretty(test)
	assert.Equal(t, prettyTestWant, got1)

	got2 := Pretty(jsonString)
	assert.Equal(t, prettyTestWant, got2)

	test3 := []byte("<fff> not a json object")
	got3 := Pretty(test3)
	assert.Equal(t, string(test3), got3)

	test4 := []byte{
		255, 253, 189, 240, 128, 200, 202, 204,
	}
	got4 := Pretty(test4)
	assert.Equal(t, "<pretty: non-printable bytes of length 8>", got4)

	got5 := Pretty2(map[string]any{"1": 123, "b": "<html>"})
	want5 := "{\n  \"1\": 123,\n  \"b\": \"<html>\"\n}"
	assert.Equal(t, want5, got5)
}
