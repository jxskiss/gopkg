// +build jsoniter

package json

import (
	"io"

	"github.com/gobwas/pool/pbytes"
	"github.com/json-iterator/go"
)

var (
	// ConfigCompatibleWithStandardLibrary tries to be 100% compatible with standard library behavior.
	cfg = jsoniter.ConfigCompatibleWithStandardLibrary
)

func Marshal(v interface{}) ([]byte, error) {
	stream := cfg.BorrowStream(nil)
	defer cfg.ReturnStream(stream)
	stream.WriteVal(v)
	if stream.Error != nil {
		return nil, stream.Error
	}
	result := stream.Buffer()
	b := pbytes.GetLen(len(result))
	copy(b, result)
	return b, nil
}

func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	return cfg.MarshalIndent(v, prefix, indent)
}

func Unmarshal(data []byte, v interface{}) error {
	return cfg.Unmarshal(data, v)
}

func NewEncoder(writer io.Writer) *jsoniter.Encoder {
	return cfg.NewEncoder(writer)
}

func NewDecoder(reader io.Reader) *jsoniter.Decoder {
	return cfg.NewDecoder(reader)
}
