package json

import (
	"bytes"
	"encoding/json"
	"io"
	"os"

	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
)

// Marshaler is an alias name of encoding/json.Marshaler.
// See encoding/json.Marshaler for detailed document.
type Marshaler = json.Marshaler

// Unmarshaler is an alias name of encoding/json.Unmarshaler.
// See encoding/json.Unmarshaler for detailed document.
type Unmarshaler = json.Unmarshaler

// Encoder is a wrapper of encoding/json.Encoder.
// It provides same methods as encoding/json.Encoder but with method
// chaining capabilities.
//
// See encoding/json.Encoder for detailed document.
type Encoder struct {
	*aliasEncoder
	disableMapOrdering bool
}

// NewEncoder returns a new Encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{aliasEncoder: _NewEncoder(w)}
}

// SetEscapeHTML specifies whether problematic HTML characters
// should be escaped inside JSON quoted strings.
// The default behavior is to escape &, <, and > to \u0026, \u003c, and \u003e
// to avoid certain safety problems that can arise when embedding JSON in HTML.
//
// In non-HTML settings where the escaping interferes with the readability
// of the output, SetEscapeHTML(false) disables this behavior.
func (enc *Encoder) SetEscapeHTML(on bool) *Encoder {
	enc.aliasEncoder.SetEscapeHTML(on)
	return enc
}

// SetIndent instructs the encoder to format each subsequent encoded
// value as if indented by the package-level function Indent(dst, src, prefix, indent).
// Calling SetIndent("", "") disables indentation.
func (enc *Encoder) SetIndent(prefix, indent string) *Encoder {
	enc.aliasEncoder.SetIndent(prefix, indent)
	return enc
}

// DisableMapOrdering instructs the encoder to not sort map keys,
// which makes it faster than default.
//
// This option has effect only when build with tag "unsafejson",
// else calling it is a no-op.
func (enc *Encoder) DisableMapOrdering() *Encoder {
	enc.disableMapOrdering = true
	return enc
}

// Encode writes the JSON encoding of v to the stream, followed by a
// newline character.
//
// See the documentation for Marshal for details about the conversion
// of Go values to JSON.
func (enc *Encoder) Encode(v interface{}) error {
	return _encoderEncode(enc.aliasEncoder, enc.disableMapOrdering, v)
}

// Decoder is a wrapper of encoding/json.Decoder.
// It provides same methods as encoding/json.Decoder but with method
// chaining capabilities.
//
// See encoding/json.Decoder for detailed document.
type Decoder struct {
	*aliasDecoder
}

// NewDecoder returns a new Decoder that reads from r.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{_NewDecoder(r)}
}

// UseNumber causes the Decoder to unmarshal a number into an interface{}
// as a Number instead of as a float64.
func (dec *Decoder) UseNumber() *Decoder {
	dec.aliasDecoder.UseNumber()
	return dec
}

// DisallowUnknownFields causes the Decoder to return an error when the
// destination is a struct and the input contains object keys which do
// not match any non-ignored, exported fields in the destination.
func (dec *Decoder) DisallowUnknownFields() *Decoder {
	dec.aliasDecoder.DisallowUnknownFields()
	return dec
}

// Marshal returns the JSON encoding of v.
//
// See encoding/json.Marshal for detailed document.
func Marshal(v interface{}) ([]byte, error) {
	return _Marshal(v)
}

// MarshalToString returns the JSON encoding of v as string.
//
// See encoding/json.Marshal for detailed document.
func MarshalToString(v interface{}) (string, error) {
	buf, err := _Marshal(v)
	if err != nil {
		return "", err
	}
	return unsafeheader.BytesToString(buf), nil
}

// MarshalIndent is like Marshal but applies Indent to format the output.
//
// See encoding/json.MarshalIndent for detailed document.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return _MarshalIndent(v, prefix, indent)
}

// MarshalNoMapOrdering is like Marshal but does not sort map keys.
// It's useful to optimize performance where map key ordering is not needed.
//
// It has effect only when build with tag "unsafejson", else it is
// an alias name of Marshal.
func MarshalNoMapOrdering(v interface{}) ([]byte, error) {
	return _MarshalNoMapOrdering(v)
}

// MarshalNoHTMLEscape is like Marshal but does not escape HTML characters.
// Optionally indent can be applied to the output,
// empty prefix and indent disables indentation.
// The output is more friendly to read for log messages.
func MarshalNoHTMLEscape(v interface{}, prefix, indent string) ([]byte, error) {
	var buf bytes.Buffer
	err := NewEncoder(&buf).
		SetEscapeHTML(false).
		SetIndent(prefix, indent).
		Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Unmarshal parses the JSON-encoded data and stores the result
// in the value pointed to by v.
//
// See encoding/json.Unmarshal for detailed document.
func Unmarshal(data []byte, v interface{}) error {
	return _Unmarshal(data, v)
}

// UnmarshalFromString parses the JSON-encoded string data and stores
// the result in the value pointed to by v.
//
// See encoding/json.Unmarshal for detailed document.
func UnmarshalFromString(data string, v interface{}) error {
	buf := unsafeheader.StringToBytes(data)
	return _Unmarshal(buf, v)
}

// Load reads JSON-encoded data from the named file at path and stores
// the result in the value pointed to by v.
func Load(path string, v interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	err = NewDecoder(file).Decode(v)
	return err
}

// Dump writes v to the named file at path using JSON encoding.
// It disables HTMLEscape.
// Optionally indent can be applied to the output,
// empty prefix and indent disables indentation.
// The output is friendly to read by humans.
func Dump(path string, v interface{}, prefix, indent string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	err = NewEncoder(file).
		SetEscapeHTML(false).
		SetIndent(prefix, indent).
		Encode(v)
	return err
}
