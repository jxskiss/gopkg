// +build jsoniter

package json

import (
	stdjson "encoding/json"
	"github.com/json-iterator/go"
)

var (
	// ConfigCompatibleWithStandardLibrary tries to be 100% compatible with standard library behavior.
	cfg = jsoniter.ConfigCompatibleWithStandardLibrary
)

var (
	Marshal       = cfg.Marshal
	MarshalIndent = cfg.MarshalIndent
	Unmarshal     = cfg.Unmarshal

	MarshalToString = cfg.MarshalToString

	NewEncoder = cfg.NewEncoder
	NewDecoder = cfg.NewDecoder

	Compact    = stdjson.Compact
	HTMLEscape = stdjson.HTMLEscape
	Indent     = stdjson.Indent
	Valid      = cfg.Valid
)

func UnmarshalFromString(str string, v interface{}) error {
	data := s2b(str)
	return cfg.Unmarshal(data, v)
}
