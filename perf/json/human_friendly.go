package json

import (
	"bytes"
	"encoding/json"
	"io"

	jsoniter "github.com/json-iterator/go"

	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
)

// HumanFriendly is a config which generates data that is more friendly
// for human reading.
// Also, this config can encode data with `interface{}` as map keys,
// in contrast, the standard library fails in this case.
var HumanFriendly = struct {
	Marshal             func(v any) ([]byte, error)
	MarshalToString     func(v any) (string, error)
	MarshalIndent       func(v any, prefix, indent string) ([]byte, error)
	MarshalIndentString func(v any, prefix, indent string) (string, error)
	NewEncoder          func(w io.Writer) *Encoder
}{
	Marshal:             hFriendlyMarshal,
	MarshalToString:     hFriendlyMarshalToString,
	MarshalIndent:       hFriendlyMarshalIndent,
	MarshalIndentString: hFriendlyMarshalIndentString,
	NewEncoder:          newHumanFriendlyEncoder,
}

var jsoniterHumanFriendlyConfig = jsoniter.Config{
	EscapeHTML:                    false,
	MarshalFloatWith6Digits:       true,
	SortMapKeys:                   true,
	UseNumber:                     true,
	ObjectFieldMustBeSimpleString: true,
}.Froze()

func hFriendlyMarshal(v any) ([]byte, error) {
	return jsoniterHumanFriendlyConfig.Marshal(v)
}

func hFriendlyMarshalToString(v any) (string, error) {
	return jsoniterHumanFriendlyConfig.MarshalToString(v)
}

func hFriendlyMarshalIndent(v any, prefix, indent string) ([]byte, error) {
	b, err := jsoniterHumanFriendlyConfig.Marshal(v)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	err = json.Indent(&buf, b, prefix, indent)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func hFriendlyMarshalIndentString(v any, prefix, indent string) (string, error) {
	buf, err := hFriendlyMarshalIndent(v, prefix, indent)
	if err != nil {
		return "", err
	}
	return unsafeheader.BytesToString(buf), nil
}

func newHumanFriendlyEncoder(w io.Writer) *Encoder {
	return &Encoder{&hFriendlyEncoder{w: w}}
}

type hFriendlyEncoder struct {
	w      io.Writer
	prefix string
	indent string
}

func (h *hFriendlyEncoder) Encode(val any) error {
	var buf []byte
	var err error
	if h.prefix == "" && h.indent == "" {
		buf, err = jsoniterHumanFriendlyConfig.Marshal(val)
	} else {
		buf, err = hFriendlyMarshalIndent(val, h.prefix, h.indent)
	}
	if err != nil {
		return err
	}
	_, err = h.w.Write(buf)
	return err
}

func (h *hFriendlyEncoder) SetEscapeHTML(_ bool) {
	return
}

func (h *hFriendlyEncoder) SetIndent(prefix, indent string) {
	h.prefix = prefix
	h.indent = indent
}
