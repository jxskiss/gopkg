package easy

import (
	"bytes"
	"github.com/jxskiss/gopkg/json"
	"io"
	"io/ioutil"
	"strings"
)

func SlashJoin(path ...string) string {
	if len(path) == 0 {
		return ""
	}
	result := path[0]
	for _, next := range path[1:] {
		aslash := strings.HasSuffix(result, "/")
		bslash := strings.HasPrefix(next, "/")
		switch {
		case aslash && bslash:
			result += next[1:]
		case !aslash && !bslash:
			result += "/" + next
		default:
			result += next
		}
	}
	return result
}

func JsonToReader(obj interface{}) (io.Reader, error) {
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(b), nil
}

func DecodeJson(r io.Reader, v interface{}) error {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
