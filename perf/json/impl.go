package json

import (
	"bytes"
	"io"
	"sync/atomic"
)

// Implementation is the interface of an underlying JSON implementation.
// User can change the underlying implementation on-the-fly at runtime.
type Implementation interface {
	Marshal(v any) ([]byte, error)
	MarshalIndent(v any, prefix, indent string) ([]byte, error)
	Unmarshal(data []byte, v any) error
	Valid(data []byte) bool

	MarshalToString(v any) (string, error)
	UnmarshalFromString(data string, v any) error

	Compact(dst *bytes.Buffer, src []byte) error
	HTMLEscape(dst *bytes.Buffer, src []byte)
	Indent(dst *bytes.Buffer, src []byte, prefix, indent string) error

	MarshalFastest(v any) ([]byte, error)
	MarshalNoHTMLEscape(v any, prefix, indent string) ([]byte, error)

	NewEncoder(w io.Writer) UnderlyingEncoder
	NewDecoder(r io.Reader) UnderlyingDecoder
}

type UnderlyingEncoder interface {
	Encode(val interface{}) error
	SetEscapeHTML(on bool)
	SetIndent(prefix, indent string)
}

type UnderlyingDecoder interface {
	Decode(val interface{}) error
	Buffered() io.Reader
	DisallowUnknownFields()
	More() bool
	UseNumber()
}

var globalImpl atomic.Pointer[Implementation]

func init() {
	globalImpl.Store(&DefaultJSONIteratorImpl)
}

func getImpl() Implementation {
	return *globalImpl.Load()
}

// ChangeImpl changes the underlying JSON implementation on-the-fly
// at runtime.
//
// You may see github.com/jxskiss/gopkg/_examples/perf/json/bytedance_sonic
// for an example to use github.com/bytedance/sonic as the underlying
// implementation.
func ChangeImpl(impl Implementation) {
	globalImpl.Store(&impl)
}
