// +build no_jsoniter

package json

import "encoding/json"

var (
	Marshal   = json.Marshal
	Unmarshal = json.Unmarshal

	NewEncoder = json.NewEncoder
	NewDecoder = json.NewDecoder
)
