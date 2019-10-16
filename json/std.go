// +build !jsoniter

package json

import (
	"encoding/json"
)

var (
	Marshal       = json.Marshal
	MarshalIndent = json.MarshalIndent
	Unmarshal     = json.Unmarshal

	NewEncoder = json.NewEncoder
	NewDecoder = json.NewDecoder

	Compact    = json.Compact
	Indent     = json.Indent
	HTMLEscape = json.HTMLEscape
	Valid      = json.Valid
)

func MarshalToString(v interface{}) (string, error) {
	buf, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return b2s(buf), nil
}

func UnmarshalFromString(str string, v interface{}) error {
	data := s2b(str)
	return json.Unmarshal(data, v)
}
