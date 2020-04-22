package json

import (
	"encoding/json"
	"github.com/json-iterator/go"
)

// ConfigCompatibleWithStandardLibrary tries to be 100% compatible with standard library behavior.
var cfg = jsoniter.ConfigCompatibleWithStandardLibrary

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
	UnsupportedTypeError  = json.UnmarshalTypeError
	UnsupportedValueError = json.UnsupportedValueError

	Any = jsoniter.Any
)

type (
	Marshaler   = json.Marshaler
	Unmarshaler = json.Unmarshaler
)

var (
	Compact    = json.Compact
	HTMLEscape = json.HTMLEscape
	Indent     = json.Indent
)

func Marshal(v interface{}) ([]byte, error) {
	switch v := v.(type) {
	case string:
		buf := make([]byte, 0, 4+len(v))
		buf = AppendString(buf, v)
		return buf, nil
	case map[string]string:
		return MarshalStringMap(v)
	default:
		return _Marshal(v)
	}
}

func MarshalToString(v interface{}) (string, error) {
	buf, err := Marshal(v)
	if err != nil {
		return "", err
	}
	return b2s(buf), nil
}

func UnmarshalFromString(str string, v interface{}) error {
	data := s2b(str)
	return Unmarshal(data, v)
}

func Get(data []byte, path ...interface{}) Any {
	return cfg.Get(data, path...)
}
