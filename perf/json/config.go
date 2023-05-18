package json

import (
	"bytes"
	std "encoding/json"
	"io"
	"log"

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

// HumanFriendly is a config which generates data that is more friendly
// for human reading.
// Also, this config can encode data with `interface{}` as map keys,
// in contrast, the standard library fails in this case.
var HumanFriendly struct {
	Marshal             func(v any) ([]byte, error)
	MarshalToString     func(v any) (string, error)
	MarshalIndent       func(v any, prefix, indent string) ([]byte, error)
	MarshalIndentString func(v any, prefix, indent string) (string, error)
	NewEncoder          func(w io.Writer) *Encoder
}

var _J apiProxy

var (
	sonicDefault    = sonic.ConfigStd
	jsoniterDefault = jsoniter.ConfigCompatibleWithStandardLibrary
)

func init() {
	// bytedance/sonic still has some bugs, which gives incorrect
	// marshaling/unmarshalling result silently in some corner case,
	// seems that it is not fully ready for being the default choice
	// for production, thus we change to use jsoniter as the default.
	//
	// Venturous user may use Config to switch to bytedance/sonic
	// explicitly, be careful.
	//
	// My may change to bytedance/sonic as default in the future,
	// when it's fully ready for production deployment.

	//{
	//	if isSonicJIT {
	//		_J.useSonicConfig(sonicDefault)
	//	} else {
	//		_J.useJSONIterConfig(jsoniterDefault)
	//	}
	//}
	_J.useJSONIterConfig(jsoniterDefault)

	HumanFriendly.Marshal = hFriendlyMarshal
	HumanFriendly.MarshalToString = hFriendlyMarshalToString
	HumanFriendly.MarshalIndent = hFriendlyMarshalIndent
	HumanFriendly.MarshalIndentString = hFriendlyMarshalIndentString
	HumanFriendly.NewEncoder = func(w io.Writer) *Encoder {
		return &Encoder{&hFriendlyEncoder{w: w}}
	}
}

type apiProxy struct {
	Marshal       func(v any) ([]byte, error)
	MarshalIndent func(v any, prefix, indent string) ([]byte, error)
	Unmarshal     func(data []byte, v any) error
	Valid         func(data []byte) bool

	MarshalToString     func(v any) (string, error)
	UnmarshalFromString func(data string, v any) error

	Compact    func(dst *bytes.Buffer, src []byte) error
	HTMLEscape func(dst *bytes.Buffer, src []byte)
	Indent     func(dst *bytes.Buffer, src []byte, prefix, indent string) error

	MarshalNoMapOrdering func(v any) ([]byte, error)
	MarshalNoHTMLEscape  func(v any, prefix, indent string) ([]byte, error)

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
	if !isSonicJIT {
		log.Println("[WARN] json: bytedance/sonic is not supported, fallback to jsoniterDefault")
		p.useJSONIterConfig(jsoniterDefault)
		return
	}
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
