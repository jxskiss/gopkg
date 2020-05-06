// +build jsoniter

package json

import "github.com/json-iterator/go"

type (
	Encoder = jsoniter.Encoder
	Decoder = jsoniter.Decoder
)

var (
	_Marshal       = cfg.Marshal
	_MarshalFast   = jsoniter.ConfigDefault.Marshal
	_MarshalIndent = cfg.MarshalIndent
	_Unmarshal     = cfg.Unmarshal
	Valid          = cfg.Valid

	NewEncoder = cfg.NewEncoder
	NewDecoder = cfg.NewDecoder
)
