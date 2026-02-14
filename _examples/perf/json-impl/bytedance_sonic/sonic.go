package bytedance_sonic

import (
	"bytes"
	"io"

	"github.com/bytedance/sonic"

	"github.com/jxskiss/gopkg/v2/perf/json"
)

// Default uses [sonic.ConfigDefault] as the underlying implementation.
var Default = New(sonic.ConfigDefault, true)

// New creates a json.Implementation based on github.com/bytedance/sonic.
// If useConfigFastest is true, it uses [sonic.ConfigFastest]
// for method MarshalFastest, else it uses api.Marshal.
func New(api sonic.API, useConfigFastest bool) json.Implementation {
	impl := &sonicImpl{
		api:            api,
		marshalFastest: api.Marshal,
	}
	if useConfigFastest {
		impl.marshalFastest = sonic.ConfigFastest.Marshal
	}
	return impl
}

type sonicImpl struct {
	api            sonic.API
	marshalFastest func(v any) ([]byte, error)
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
	return impl.marshalFastest(v)
}

func (impl sonicImpl) MarshalNoHTMLEscape(v any, prefix, indent string) ([]byte, error) {
	return json.StdImpl.MarshalNoHTMLEscape(v, prefix, indent)
}

func (impl sonicImpl) NewEncoder(w io.Writer) *json.Encoder {
	return &json.Encoder{UnderlyingEncoder: impl.api.NewEncoder(w)}
}

func (impl sonicImpl) NewDecoder(r io.Reader) *json.Decoder {
	return &json.Decoder{UnderlyingDecoder: impl.api.NewDecoder(r)}
}
