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
	NewEncoder:          hFriendlyNewEncoder,
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

// 注意 jsoniter 对 prefix, indent 的处理有问题，这里不能直接使用
// jsoniterHumanFriendlyConfig.NewEncoder
func hFriendlyNewEncoder(w io.Writer) *Encoder {
	buf := bytes.NewBuffer(nil)
	return &Encoder{&hFriendlyEncoder{
		w:   w,
		buf: buf,
		enc: jsoniterHumanFriendlyConfig.NewEncoder(buf),
	}}
}

type hFriendlyEncoder struct {
	w      io.Writer
	buf    *bytes.Buffer
	enc    *jsoniter.Encoder
	prefix string
	indent string
}

func (h *hFriendlyEncoder) Encode(val any) error {
	var err error
	var out []byte
	err = h.enc.Encode(val)
	if err != nil {
		return err
	}
	if h.prefix != "" || h.indent != "" {
		var indentBuf bytes.Buffer
		err = json.Indent(&indentBuf, h.buf.Bytes(), h.prefix, h.indent)
		if err != nil {
			return err
		}
		out = indentBuf.Bytes()
	} else {
		out = h.buf.Bytes()
	}
	_, err = h.w.Write(out)
	return err
}

func (h *hFriendlyEncoder) SetEscapeHTML(on bool) {
	h.enc.SetEscapeHTML(on)
}

func (h *hFriendlyEncoder) SetIndent(prefix, indent string) {
	h.prefix = prefix
	h.indent = indent
}
