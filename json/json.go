package json

import (
	"encoding/json"
	"github.com/json-iterator/go"
	"os"
	"reflect"
	"strconv"
	"strings"
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
		return marshalOptimized(v, AppendIntSlice)
	case isStringSlice(typ):
		return marshalOptimized(v, appendStringSlice)
	default:
		return _Marshal(v)
	}
}

func MarshalFast(v interface{}) ([]byte, error) {
	ok, buf, err := marshalNilOrMarshaler(v)
	if ok {
		return buf, err
	}
	typ := reflect.TypeOf(v)
	if isOptimizedType(typ) {
		appendFunc := getAppendFunc(typ)
		return marshalOptimized(v, appendFunc)
	}
	return _MarshalFast(v)
}

func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return _MarshalIndent(v, prefix, indent)
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

func GetByDot(data []byte, path string) Any {
	return cfg.Get(data, splitDotPath(path)...)
}

func splitDotPath(path string) []interface{} {
	parts := strings.Split(path, ".")
	out := make([]interface{}, 0, len(parts))
	for _, s := range parts {
		switch {
		case isDigits(s):
			idx, _ := strconv.ParseInt(s, 10, 64)
			out = append(out, int(idx))
		case s == "*":
			out = append(out, '*')
		default:
			out = append(out, s)
		}
	}
	return out
}

func isDigits(s string) bool {
	for _, x := range s {
		if !('0' <= x && x <= '9') {
			return false
		}
	}
	return true
}

func Load(path string, v interface{}) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	err = NewDecoder(file).Decode(v)
	return err
}

func Dump(path string, v interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	err = NewEncoder(file).Encode(v)
	return err
}
