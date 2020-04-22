// +build jsoniter

package json

import "github.com/json-iterator/go"

type (
	Encoder = jsoniter.Encoder
	Decoder = jsoniter.Decoder
)

var (
	_Marshal      = cfg.Marshal
	MarshalIndent = cfg.MarshalIndent
	Unmarshal     = cfg.Unmarshal
	Valid         = cfg.Valid

	NewEncoder = cfg.NewEncoder
	NewDecoder = cfg.NewDecoder
)
