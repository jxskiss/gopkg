package json

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/jxskiss/gopkg/v2/internal/unsafeheader"
)

func stdMarshalToString(v interface{}) (string, error) {
	buf, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return unsafeheader.BytesToString(buf), nil
}

func stdUnmarshalFromString(data string, v interface{}) error {
	buf := unsafeheader.StringToBytes(data)
	return json.Unmarshal(buf, v)
}

func stdMarshalNoHTMLEscape(v interface{}, prefix, indent string) ([]byte, error) {
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

func stdNewEncoder(w io.Writer) underlyingEncoder {
	return json.NewEncoder(w)
}

func stdNewDecoder(r io.Reader) underlyingDecoder {
	return json.NewDecoder(r)
}
