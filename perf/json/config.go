package json

import (
	"bytes"
	std "encoding/json"
	"io"

	"github.com/bytedance/sonic"
)

type Options struct {
	DisableSonic bool
	SonicConfig  sonic.API
}

// Config configures the behavior of this json package.
//
// It may be called at program startup, library code shall not call
// this function.
func Config(opts Options) {
	if opts.DisableSonic {
		_J.Marshal = std.Marshal
		_J.MarshalIndent = std.MarshalIndent
		_J.Unmarshal = std.Unmarshal
		_J.Valid = std.Valid

		_J.MarshalToString = stdMarshalToString
		_J.UnmarshalFromString = stdUnmarshalFromString

		_J.Compact = std.Compact
		_J.HTMLEscape = std.HTMLEscape
		_J.Indent = std.Indent

		_J.MarshalNoMapOrdering = std.Marshal
		_J.MarshalNoHTMLEscape = stdMarshalNoHTMLEscape

		_J.NewEncoder = stdNewEncoder
		_J.NewDecoder = stdNewDecoder

	} else if opts.SonicConfig != nil {
		cfg := opts.SonicConfig
		_J.Marshal = cfg.Marshal
		_J.MarshalIndent = cfg.MarshalIndent
		_J.Unmarshal = cfg.Unmarshal
		_J.Valid = cfg.Valid

		_J.MarshalToString = cfg.MarshalToString
		_J.UnmarshalFromString = cfg.UnmarshalFromString

		_J.Compact = std.Compact
		_J.HTMLEscape = std.HTMLEscape
		_J.Indent = std.Indent

		_J.MarshalNoMapOrdering = sonicMarshalNoMapOrdering(cfg)
		_J.MarshalNoHTMLEscape = sonicMarshalNoHTMLEscape(cfg)

		_J.NewEncoder = sonicNewEncoder(cfg)
		_J.NewDecoder = sonicNewDecoder(cfg)
	}
}

var _J = struct {
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
}{
	Marshal:       sonicDefault.Marshal,
	MarshalIndent: sonicDefault.MarshalIndent,
	Unmarshal:     sonicDefault.Unmarshal,
	Valid:         sonicDefault.Valid,

	MarshalToString:     sonicDefault.MarshalToString,
	UnmarshalFromString: sonicDefault.UnmarshalFromString,

	Compact:    std.Compact,
	HTMLEscape: std.HTMLEscape,
	Indent:     std.Indent,

	MarshalNoMapOrdering: sonicMarshalNoMapOrdering(sonicDefault),
	MarshalNoHTMLEscape:  sonicMarshalNoHTMLEscape(sonicDefault),

	NewEncoder: sonicNewEncoder(sonicDefault),
	NewDecoder: sonicNewDecoder(sonicDefault),
}
