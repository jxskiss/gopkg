package json

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
)

// StdImpl uses package "encoding/json" in the standard library
// as the underlying implementation.
var StdImpl Implementation = stdImpl{}

type stdImpl struct{}

func (stdImpl) Marshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (stdImpl) MarshalIndent(v any, prefix, indent string) ([]byte, error) {
	return json.MarshalIndent(v, prefix, indent)
}

func (stdImpl) Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func (stdImpl) Valid(data []byte) bool {
	return json.Valid(data)
}

func (stdImpl) MarshalToString(v any) (string, error) {
	buf, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return unsafeheader.BytesToString(buf), nil
}

func (stdImpl) UnmarshalFromString(data string, v any) error {
	buf := unsafeheader.StringToBytes(data)
	return json.Unmarshal(buf, v)
}

func (stdImpl) Compact(dst *bytes.Buffer, src []byte) error {
	return json.Compact(dst, src)
}

func (stdImpl) HTMLEscape(dst *bytes.Buffer, src []byte) {
	json.HTMLEscape(dst, src)
}

func (stdImpl) Indent(dst *bytes.Buffer, src []byte, prefix, indent string) error {
	return json.Indent(dst, src, prefix, indent)
}

func (stdImpl) MarshalFastest(v any) ([]byte, error) {
	return json.Marshal(v)
}

func (stdImpl) MarshalNoHTMLEscape(v any, prefix, indent string) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent(prefix, indent)
	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}

	// json.Encoder always appends '\n' after encoding,
	// which is not same with json.Marshal.
	out := buf.Bytes()
	if len(out) > 0 && out[len(out)-1] == '\n' {
		out = out[:len(out)-1]
	}
	return out, nil
}

func (stdImpl) NewEncoder(w io.Writer) UnderlyingEncoder {
	return json.NewEncoder(w)
}

func (stdImpl) NewDecoder(r io.Reader) UnderlyingDecoder {
	return json.NewDecoder(r)
}
