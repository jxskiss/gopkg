// +build !jsoniter

package json

import "encoding/json"

type (
	Encoder = json.Encoder
	Decoder = json.Decoder
)

var (
	Marshal       = json.Marshal
	MarshalIndent = json.MarshalIndent
	Unmarshal     = json.Unmarshal
	Valid         = json.Valid

	NewEncoder = json.NewEncoder
	NewDecoder = json.NewDecoder
)
