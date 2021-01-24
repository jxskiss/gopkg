// +build jsoniter

package json

import "github.com/json-iterator/go"

type (
	Encoder = jsoniter.Encoder
	Decoder = jsoniter.Decoder
)

var (
	_Marshal       = stdcfg.Marshal
	_MarshalIndent = stdcfg.MarshalIndent
	_Unmarshal     = stdcfg.Unmarshal
	Valid          = stdcfg.Valid

	NewEncoder = stdcfg.NewEncoder
	NewDecoder = stdcfg.NewDecoder
)
