package json

import (
	"reflect"
	"testing"
)

var malformedJSONData = `
{
	// A comment! You normally can't put these in JSON
	"obj1": {
		"foo": "bar", // <-- A trailing comma! No worries.
	},
	/*
	This style of comments will also be safely removed.
	*/
	"array": [1, 2, 3, ], // Trailing comma in array.
	"import": @import("testdata.json"), // Import another json file.
	"obj2": {
		"foo": "bar", /* Another style inline comment. */
	}, // <-- Another trailing comma!
}
`

func TestUnmarshalExt_comment_trailingComma(t *testing.T) {
	want := map[string]interface{}{
		"obj1": map[string]interface{}{
			"foo": "bar",
		},
		"array": []interface{}{float64(1), float64(2), float64(3)},
		"import": map[string]interface{}{
			"foo": "bar",
		},
		"obj2": map[string]interface{}{
			"foo": "bar",
		},
	}
	got := make(map[string]interface{})
	err := UnmarshalExt([]byte(malformedJSONData), &got, "")
	if err != nil {
		t.Fatalf("failed unmarshal malformed json: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expecting equal: got = %v, want = %v", got, want)
	}
}

func TestUnmarshalExt_UnicodeEscape(t *testing.T) {
	jsonData := `["Grammar \u0026 Mechanics \/ Word Work"]`
	got := make([]string, 0)
	err := UnmarshalExt([]byte(jsonData), &got, "")
	if err != nil {
		t.Errorf("failed unmarshal unicode escape char: %v", err)
	}
}

func TestUnmarshalExt_SingleQuote(t *testing.T) {
	jsonData := `{'key\'': 'value"'}`
	got := make(map[string]string)
	err := UnmarshalExt([]byte(jsonData), &got, "")
	if err != nil {
		t.Errorf("failed unmarshal single quoted string: %v", err)
	}
	if got["key'"] != "value\"" {
		t.Errorf("unmarshal single quoted string: incorrect key value")
	}
}
