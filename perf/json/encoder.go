package json

import "io"

// Encoder is a wrapper of encoding/json.Encoder.
// It provides same methods as encoding/json.Encoder but with method
// chaining capabilities.
//
// See encoding/json.Encoder for detailed document.
type Encoder struct {
	UnderlyingEncoder
}

// NewEncoder returns a new Encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	return getImpl().NewEncoder(w)
}

// SetEscapeHTML specifies whether problematic HTML characters
// should be escaped inside JSON quoted strings.
// The default behavior is to escape &, <, and > to \u0026, \u003c, and \u003e
// to avoid certain safety problems that can arise when embedding JSON in HTML.
//
// In non-HTML settings where the escaping interferes with the readability
// of the output, SetEscapeHTML(false) disables this behavior.
func (enc *Encoder) SetEscapeHTML(on bool) *Encoder {
	enc.UnderlyingEncoder.SetEscapeHTML(on)
	return enc
}

// SetIndent instructs the encoder to format each subsequent encoded
// value as if indented by the package-level function Indent(dst, src, prefix, indent).
// Calling SetIndent("", "") disables indentation.
func (enc *Encoder) SetIndent(prefix, indent string) *Encoder {
	enc.UnderlyingEncoder.SetIndent(prefix, indent)
	return enc
}
