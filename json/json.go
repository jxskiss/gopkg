package json

import (
	"encoding/json"
	"github.com/json-iterator/go"
	"reflect"
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
	ok, buf, err := marshalNilOrMarshaler(v)
	if ok {
		return buf, err
	}
	typ := reflect.TypeOf(v)
	switch {
	case isIntSlice(typ):
		return marshalIntSlice(v)
	case isStringSlice(typ):
		slice := castStringSlice(v)
		return marshalStringSlice(slice)
	case isStringMap(typ):
		strMap := castStringMap(v)
		return marshalStringMap(strMap)
	case isStringInterfaceMap(typ):
		strMap := castStringInterfaceMap(v)
		return marshalStringInterfaceMap(strMap)
	default:
		return _Marshal(v)
	}
}

func Unmarshal(data []byte, v interface{}) error {
	if isUnmarshaler(v) {
		return _Unmarshal(data, v)
	}
	typ := reflect.TypeOf(v)
	if isStringMapPtr(typ) {
		ptr := castStringMapPtr(v)
		return unmarshalStringMap(data, ptr)
	}
	return _Unmarshal(data, v)
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
