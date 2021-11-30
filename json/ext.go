package json

import (
	"github.com/jxskiss/gopkg/v2/json/extparser"
	"io/ioutil"
	"os"
)

// UnmarshalExt parses the JSON-encoded data and stores the result in the
// value pointed to by v.
//
// In addition to features of encoding/json, it enables some extended
// features such as "trailing comma", "comments", "file including", etc.
// The extended features are documented in the README file.
func UnmarshalExt(data []byte, v interface{}, importRoot string) error {
	if importRoot == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		importRoot = wd
	}
	data, err := extparser.Parse(data, importRoot)
	if err != nil {
		return err
	}
	return Unmarshal(data, v)
}

// LoadExt reads JSON-encoded data from the named file at path and stores
// the result in the value pointed to by v.
//
// In additional to features of encoding/json, it enables some extended
// features such as "trailing comma", "comments", "file including" etc.
// The extended features are documented in the README file.
func LoadExt(path string, v interface{}, importRoot string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	return UnmarshalExt(data, v, importRoot)
}
