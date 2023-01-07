package json

import (
	"bytes"
	std "encoding/json"
	"io"

	"github.com/bytedance/sonic"
	jsoniter "github.com/json-iterator/go"
)

type Options struct {
	UseStdlib         bool
	UseJSONIterConfig jsoniter.API
	UseSonicConfig    sonic.API
}

// Config configures the behavior of this json package.
//
// It may be called at program startup, library code shall not call
// this function.
func Config(opts Options) {
	if opts.UseStdlib {
		_J.useStdlib()
	} else if opts.UseJSONIterConfig != nil {
		_J.useJSONIterConfig(opts.UseJSONIterConfig)
	} else if opts.UseSonicConfig != nil {
		_J.useSonicConfig(opts.UseSonicConfig)
	}
}

var _J apiProxy

func init() {
	_J.useSonicConfig(sonicDefault)
}

type apiProxy struct {
	Marshal       func(v interface{}) ([]byte, error)
	MarshalIndent func(v interface{}, prefix, indent string) ([]byte, error)
	Unmarshal     func(data []byte, v interface{}) error
	Valid         func(data []byte) bool

	MarshalToString     func(v interface{}) (string, error)
	UnmarshalFromString func(data string, v interface{}) error

	Compact    func(dst *bytes.Buffer, src []byte) error
	HTMLEscape func(dst *bytes.Buffer, src []byte)
	Indent     func(dst *bytes.Buffer, src []byte, prefix, indent string) error

	MarshalNoMapOrdering func(v interface{}) ([]byte, error)
	MarshalNoHTMLEscape  func(v interface{}, prefix, indent string) ([]byte, error)

	NewEncoder func(w io.Writer) underlyingEncoder
	NewDecoder func(r io.Reader) underlyingDecoder
}

func (p *apiProxy) useStdlib() {
	*p = apiProxy{
		Marshal:       std.Marshal,
		MarshalIndent: std.MarshalIndent,
		Unmarshal:     std.Unmarshal,
		Valid:         std.Valid,

		MarshalToString:     stdMarshalToString,
		UnmarshalFromString: stdUnmarshalFromString,

		Compact:    std.Compact,
		HTMLEscape: std.HTMLEscape,
		Indent:     std.Indent,

		MarshalNoMapOrdering: std.Marshal,
		MarshalNoHTMLEscape:  stdMarshalNoHTMLEscape,

		NewEncoder: stdNewEncoder,
		NewDecoder: stdNewDecoder,
	}
}

func (p *apiProxy) useJSONIterConfig(api jsoniter.API) {
	*p = apiProxy{
		Marshal:       api.Marshal,
		MarshalIndent: api.MarshalIndent,
		Unmarshal:     api.Unmarshal,
		Valid:         api.Valid,

		MarshalToString:     api.MarshalToString,
		UnmarshalFromString: api.UnmarshalFromString,

		Compact:    std.Compact,
		HTMLEscape: std.HTMLEscape,
		Indent:     std.Indent,

		MarshalNoMapOrdering: jsoniterMarshalNoMapOrdering,
		MarshalNoHTMLEscape:  jsoniterMarshalNoHTMLEscape(api),

		NewEncoder: jsoniterNewEncoder(api),
		NewDecoder: jsoniterNewDecoder(api),
	}
}

func (p *apiProxy) useSonicConfig(api sonic.API) {
	*p = apiProxy{
		Marshal:       api.Marshal,
		MarshalIndent: api.MarshalIndent,
		Unmarshal:     api.Unmarshal,
		Valid:         api.Valid,

		MarshalToString:     api.MarshalToString,
		UnmarshalFromString: api.UnmarshalFromString,

		Compact:    std.Compact,
		HTMLEscape: std.HTMLEscape,
		Indent:     std.Indent,

		MarshalNoMapOrdering: sonicMarshalNoMapOrdering,
		MarshalNoHTMLEscape:  sonicMarshalNoHTMLEscape(api),

		NewEncoder: sonicNewEncoder(api),
		NewDecoder: sonicNewDecoder(api),
	}
}
