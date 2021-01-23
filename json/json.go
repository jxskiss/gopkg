package json

import (
	"encoding/json"
	"github.com/json-iterator/go"
	"os"
	"reflect"
	"unsafe"
)

// ConfigCompatibleWithStandardLibrary tries to be 100% compatible with standard library behavior.
var stdcfg = jsoniter.ConfigCompatibleWithStandardLibrary

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
	return _Marshal(v)
}

func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return _MarshalIndent(v, prefix, indent)
}

func Unmarshal(data []byte, v interface{}) error {
	if ptr, ok := v.(*map[string]string); ok {
		return UnmarshalStringMap(data, ptr)
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

func b2s(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func s2b(s string) []byte {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := &reflect.SliceHeader{
		Data: sh.Data,
		Len:  sh.Len,
		Cap:  sh.Len,
	}
	return *(*[]byte)(unsafe.Pointer(bh))
}
