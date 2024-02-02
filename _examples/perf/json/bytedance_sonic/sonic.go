package bytedance_sonic

import (
	"bytes"
	"io"

	"github.com/bytedance/sonic"

	"github.com/jxskiss/gopkg/v2/perf/json"
)

func NewSonicImpl(api sonic.API) json.Implementation {
	impl := &sonicImpl{
		api: api,
	}
	impl.marshalNoHTMLEscape = marshalNoHTMLEscape(api)
	impl.encoderFactory = newEncoderFactory(api)
	impl.decoderFactor = newDecodeFactory(api)
	return impl
}

type sonicImpl struct {
	api sonic.API

	marshalNoHTMLEscape func(v any, prefix, indent string) ([]byte, error)
	encoderFactory      func(w io.Writer) json.UnderlyingEncoder
	decoderFactor       func(r io.Reader) json.UnderlyingDecoder
}

func (impl sonicImpl) Marshal(v any) ([]byte, error) {
	return impl.api.Marshal(v)
}

func (impl sonicImpl) MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return impl.api.MarshalIndent(v, prefix, indent)
}

func (impl sonicImpl) Unmarshal(data []byte, v any) error {
	return impl.api.Unmarshal(data, v)
}

func (impl sonicImpl) Valid(data []byte) bool {
	return impl.api.Valid(data)
}

func (impl sonicImpl) MarshalToString(v any) (string, error) {
	return impl.api.MarshalToString(v)
}

func (impl sonicImpl) UnmarshalFromString(data string, v any) error {
	return impl.api.UnmarshalFromString(data, v)
}

func (impl sonicImpl) Compact(dst *bytes.Buffer, src []byte) error {
	return json.StdImpl.Compact(dst, src)
}

func (impl sonicImpl) HTMLEscape(dst *bytes.Buffer, src []byte) {
	json.StdImpl.HTMLEscape(dst, src)
}

func (impl sonicImpl) Indent(dst *bytes.Buffer, src []byte, prefix, indent string) error {
	return json.StdImpl.Indent(dst, src, prefix, indent)
}

func (impl sonicImpl) MarshalFastest(v any) ([]byte, error) {
	return marshalFastest(v)
}

func (impl sonicImpl) MarshalNoHTMLEscape(v any, prefix, indent string) ([]byte, error) {
	return impl.marshalNoHTMLEscape(v, prefix, indent)
}

func (impl sonicImpl) NewEncoder(w io.Writer) json.UnderlyingEncoder {
	return impl.encoderFactory(w)
}

func (impl sonicImpl) NewDecoder(r io.Reader) json.UnderlyingDecoder {
	return impl.decoderFactor(r)
}
