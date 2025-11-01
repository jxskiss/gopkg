// Package json provides an on-the-fly change-able API for JSON serialization.
package json

import (
	"bytes"
	"encoding/json"
)

// Marshaler is an alias name of encoding/json.Marshaler.
// See encoding/json.Marshaler for detailed document.
type Marshaler = json.Marshaler

// Unmarshaler is an alias name of encoding/json.Unmarshaler.
// See encoding/json.Unmarshaler for detailed document.
type Unmarshaler = json.Unmarshaler

// RawMessage is a raw encoded JSON value.
// It implements Marshaler and Unmarshaler and can
// be used to delay JSON decoding or precompute a JSON encoding.
type RawMessage = json.RawMessage

// Marshal returns the JSON encoding of v.
//
// See encoding/json.Marshal for detailed document.
func Marshal(v any) ([]byte, error) {
	return getImpl().Marshal(v)
}

// MarshalToString returns the JSON encoding of v as string.
//
// See encoding/json.Marshal for detailed document.
func MarshalToString(v any) (string, error) {
	return getImpl().MarshalToString(v)
}

// MarshalIndent is like Marshal but applies Indent to format the output.
//
// See encoding/json.MarshalIndent for detailed document.
func MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return getImpl().MarshalIndent(v, prefix, indent)
}

// Unmarshal parses the JSON-encoded data and stores the result
// in the value pointed to by v.
//
// See encoding/json.Unmarshal for detailed document.
func Unmarshal(data []byte, v any) error {
	return getImpl().Unmarshal(data, v)
}

// UnmarshalFromString parses the JSON-encoded string data and stores
// the result in the value pointed to by v.
//
// See encoding/json.Unmarshal for detailed document.
func UnmarshalFromString(data string, v any) error {
	return getImpl().UnmarshalFromString(data, v)
}

// Valid reports whether data is a valid JSON encoding.
func Valid(data []byte) bool {
	return getImpl().Valid(data)
}

// Compact appends to dst the JSON-encoded src with
// insignificant space characters elided.
func Compact(dst *bytes.Buffer, src []byte) error {
	return getImpl().Compact(dst, src)
}

// HTMLEscape appends to dst the JSON-encoded src with <, >, &, U+2028 and U+2029
// characters inside string literals changed to \u003c, \u003e, \u0026, \u2028, \u2029
// so that the JSON will be safe to embed inside HTML <script> tags.
// For historical reasons, web browsers don't honor standard HTML
// escaping within <script> tags, so an alternative JSON encoding must
// be used.
func HTMLEscape(dst *bytes.Buffer, src []byte) {
	getImpl().HTMLEscape(dst, src)
}

// Indent appends to dst an indented form of the JSON-encoded src.
// See encoding/json.Indent for detailed document.
func Indent(dst *bytes.Buffer, src []byte, prefix, indent string) error {
	return getImpl().Indent(dst, src, prefix, indent)
}

// MarshalFastest uses the fastest config if the underlying implementation
// supports it, e.g. jsoniter and sonic.
// The result may be incompatible with std [encoding/json] in some ways,
// especially that map keys may be not sorted.
func MarshalFastest(v any) ([]byte, error) {
	return getImpl().MarshalFastest(v)
}

// MarshalNoHTMLEscape is like Marshal but does not escape HTML characters.
// Optionally indent can be applied to the output,
// empty prefix and indent disables indentation.
// The output is more friendly to read for log messages.
func MarshalNoHTMLEscape(v any, prefix, indent string) ([]byte, error) {
	return getImpl().MarshalNoHTMLEscape(v, prefix, indent)
}
