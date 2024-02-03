package json

import (
	"bytes"
	"io"

	jsoniter "github.com/json-iterator/go"
)

// DefaultJSONIteratorImpl uses [jsoniter.ConfigCompatibleWithStandardLibrary]
// as the underlying implementation.
var DefaultJSONIteratorImpl = NewJSONIteratorImpl(
	jsoniter.ConfigCompatibleWithStandardLibrary, true)

// NewJSONIteratorImpl returns an implementation which uses api as the
// underlying config.
// If useConfigFastest is true, it uses [jsoniter.ConfigFastest]
// for method MarshalFastest, else it uses api.Marshal.
func NewJSONIteratorImpl(api jsoniter.API, useConfigFastest bool) Implementation {
	impl := &jsoniterImpl{
		api:            api,
		marshalFastest: api.Marshal,
	}
	if useConfigFastest {
		impl.marshalFastest = jsoniter.ConfigFastest.Marshal
	}
	return impl
}

type jsoniterImpl struct {
	api            jsoniter.API
	marshalFastest func(v any) ([]byte, error)
}

func (impl jsoniterImpl) Marshal(v any) ([]byte, error) {
	return impl.api.Marshal(v)
}

func (impl jsoniterImpl) MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return impl.api.MarshalIndent(v, prefix, indent)
}

func (impl jsoniterImpl) Unmarshal(data []byte, v any) error {
	return impl.api.Unmarshal(data, v)
}

func (impl jsoniterImpl) Valid(data []byte) bool {
	return impl.api.Valid(data)
}

func (impl jsoniterImpl) MarshalToString(v any) (string, error) {
	return impl.api.MarshalToString(v)
}

func (impl jsoniterImpl) UnmarshalFromString(data string, v any) error {
	return impl.api.UnmarshalFromString(data, v)
}

func (impl jsoniterImpl) Compact(dst *bytes.Buffer, src []byte) error {
	return StdImpl.Compact(dst, src)
}

func (impl jsoniterImpl) HTMLEscape(dst *bytes.Buffer, src []byte) {
	StdImpl.HTMLEscape(dst, src)
}

func (impl jsoniterImpl) Indent(dst *bytes.Buffer, src []byte, prefix, indent string) error {
	return StdImpl.Indent(dst, src, prefix, indent)
}

func (impl jsoniterImpl) MarshalFastest(v any) ([]byte, error) {
	return impl.marshalFastest(v)
}

func (impl jsoniterImpl) MarshalNoHTMLEscape(v any, prefix, indent string) ([]byte, error) {
	return StdImpl.MarshalNoHTMLEscape(v, prefix, indent)
}

func (impl jsoniterImpl) NewEncoder(w io.Writer) UnderlyingEncoder {
	return impl.api.NewEncoder(w)
}

func (impl jsoniterImpl) NewDecoder(r io.Reader) UnderlyingDecoder {
	return impl.api.NewDecoder(r)
}
