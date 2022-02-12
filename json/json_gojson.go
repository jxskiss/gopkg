//go:build unsafejson

package json

import (
	"bytes"

	"github.com/goccy/go-json"
)

type (
	Delim      = json.Delim
	Token      = json.Token
	Number     = json.Number
	RawMessage = json.RawMessage

	InvalidUTF8Error      = json.InvalidUTF8Error
	InvalidUnmarshalError = json.InvalidUnmarshalError
	MarshalerError        = json.MarshalerError
	SyntaxError           = json.SyntaxError
	UnmarshalFieldError   = json.UnmarshalFieldError
	UnmarshalTypeError    = json.UnmarshalTypeError
	UnsupportedTypeError  = json.UnsupportedTypeError
	UnsupportedValueError = json.UnsupportedValueError
)

func Compact(dst *bytes.Buffer, src []byte) error {
	return json.Compact(dst, src)
}

func HTMLEscape(dst *bytes.Buffer, src []byte) {
	json.HTMLEscape(dst, src)
}

func Indent(dst *bytes.Buffer, src []byte, prefix, indent string) error {
	return json.Indent(dst, src, prefix, indent)
}

func Valid(data []byte) bool {
	return json.Valid(data)
}

var (
	_Marshal       = json.Marshal
	_MarshalIndent = json.MarshalIndent
	_Unmarshal     = json.Unmarshal
)

func _MarshalNoMapOrdering(v interface{}) ([]byte, error) {
	return json.MarshalWithOption(v, json.UnorderedMap())
}

type (
	aliasEncoder = json.Encoder
	aliasDecoder = json.Decoder
)

var (
	_NewEncoder = json.NewEncoder
	_NewDecoder = json.NewDecoder
)

func _encoderEncode(enc *aliasEncoder, disableMapOrdering bool, v interface{}) error {
	if disableMapOrdering {
		return enc.EncodeWithOption(v, json.UnorderedMap())
	}
	return enc.Encode(v)
}
