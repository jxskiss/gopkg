package json

import (
	"encoding/json"
	"github.com/jxskiss/gopkg/bbp"
	"github.com/jxskiss/gopkg/internal/unsafeheader"
	"io"
	"os"
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
}

// NewEncoder returns a new Encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{_NewEncoder(w)}
}

func (enc *Encoder) SetEscapeHTML(on bool) *Encoder {
	enc.aliasEncoder.SetEscapeHTML(on)
	return enc
}

func (enc *Encoder) SetIndent(prefix, indent string) *Encoder {
	enc.aliasEncoder.SetIndent(prefix, indent)
	return enc
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

func (dec *Decoder) UseNumber() *Decoder {
	dec.aliasDecoder.UseNumber()
	return dec
}

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
	return unsafeheader.BtoS(buf), nil
}

// MarshalIndent is like Marshal but applies Indent to format the output.
//
// See encoding/json.MarshalIndent for detailed document.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return _MarshalIndent(v, prefix, indent)
}

// MarshalMapUnordered is like Marshal but does not sort map keys.
// It's useful to optimize performance where map key ordering is not needed.
func MarshalMapUnordered(v interface{}) ([]byte, error) {
	return _MarshalMapUnordered(v)
}

var noHTMLEscapeBufpool bbp.Pool

// MarshalNoHTMLEscape is like Marshal but does not escape HTML characters.
// Optionally indent can be applied to the output,
// empty indentPrefix and indent disables indentation.
// The output is more friendly to read for log messages.
func MarshalNoHTMLEscape(v interface{}, indentPrefix, indent string) ([]byte, error) {
	buf := noHTMLEscapeBufpool.Get()
	defer noHTMLEscapeBufpool.Put(buf)

	err := NewEncoder(buf).
		SetEscapeHTML(false).
		SetIndent(indentPrefix, indent).
		Encode(v)
	if err != nil {
		return nil, err
	}
	return buf.Copy(), nil
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
	buf := unsafeheader.StoB(data)
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
// empty indentPrefix and indent disables indentation.
func Dump(path string, v interface{}, indentPrefix, indent string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	err = NewEncoder(file).
		SetEscapeHTML(false).
		SetIndent(indentPrefix, indent).
		Encode(v)
	return err
}
