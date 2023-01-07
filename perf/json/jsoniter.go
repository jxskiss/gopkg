package json

import (
	"bytes"
	"io"

	jsoniter "github.com/json-iterator/go"
)

func jsoniterMarshalNoMapOrdering(v interface{}) ([]byte, error) {
	return jsoniter.ConfigFastest.Marshal(v)
}

func jsoniterMarshalNoHTMLEscape(cfg jsoniter.API) func(v interface{}, prefix, indent string) ([]byte, error) {
	return func(v interface{}, prefix, indent string) ([]byte, error) {
		var buf bytes.Buffer
		enc := cfg.NewEncoder(&buf)
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
}

func jsoniterNewEncoder(cfg jsoniter.API) func(w io.Writer) underlyingEncoder {
	return func(w io.Writer) underlyingEncoder {
		return cfg.NewEncoder(w)
	}
}

func jsoniterNewDecoder(cfg jsoniter.API) func(r io.Reader) underlyingDecoder {
	return func(r io.Reader) underlyingDecoder {
		return cfg.NewDecoder(r)
	}
}
