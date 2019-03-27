// +build !jsoniter

package json

import "encoding/json"

var (
	Marshal       = json.Marshal
	MarshalIndent = json.MarshalIndent
	Unmarshal     = json.Unmarshal

	NewEncoder = json.NewEncoder
	NewDecoder = json.NewDecoder
)
