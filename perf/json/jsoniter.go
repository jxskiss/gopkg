package json

import (
	"io"

	jsoniter "github.com/json-iterator/go"
)

func jsoniterMarshalFastest(v any) ([]byte, error) {
	return jsoniter.ConfigFastest.Marshal(v)
}

// MarshalNoHTMLEscape is not designed for performance critical use-case,
// we use the std [encoding/json] implementation.
func jsoniterMarshalNoHTMLEscape(_ jsoniter.API) func(v any, prefix, indent string) ([]byte, error) {
	return stdMarshalNoHTMLEscape
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
