// +build !jsoniter

package json

import "encoding/json"

type (
	Encoder = json.Encoder
	Decoder = json.Decoder
)

var (
	_Marshal      = json.Marshal
	MarshalIndent = json.MarshalIndent
	_Unmarshal    = json.Unmarshal
	Valid         = json.Valid

	NewEncoder = json.NewEncoder
	NewDecoder = json.NewDecoder
)
