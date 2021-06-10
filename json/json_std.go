// +build !gojson,!jsoniter

package json

import "encoding/json"

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

var (
	_Marshal             = json.Marshal
	_MarshalIndent       = json.MarshalIndent
	_MarshalMapUnordered = json.Marshal
	_Unmarshal           = json.Unmarshal
)

var (
	Compact    = json.Compact
	HTMLEscape = json.HTMLEscape
	Indent     = json.Indent
	Valid      = json.Valid
)

type (
	aliasEncoder = json.Encoder
	aliasDecoder = json.Decoder
)

var (
	_NewEncoder = json.NewEncoder
	_NewDecoder = json.NewDecoder
)
