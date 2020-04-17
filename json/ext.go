package json

import (
	"io"
	"io/ioutil"
	"os"

	"github.com/jxskiss/gopkg/json/extparser"
)

type ExtDecoder struct {
	reader io.Reader

	importRoot string
}

func NewExtDecoder(r io.Reader) *ExtDecoder {
	return &ExtDecoder{reader: r}
}

func (r *ExtDecoder) SetImportRoot(path string) *ExtDecoder {
	r.importRoot = path
	return r
}

func (r *ExtDecoder) Decode(v interface{}) error {
	content, err := ioutil.ReadAll(r.reader)
	if err != nil {
		return err
	}

	if r.importRoot == "" {
		wd, err := os.Getwd()
		if err != nil {
			return err
		}
		r.importRoot = wd
	}

	data, err := extparser.Parse(content, r.importRoot)
	if err != nil {
		return err
	}
	return Unmarshal(data, v)
}

func UnmarshalExt(data []byte, v interface{}) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	data, err = extparser.Parse(data, wd)
	if err != nil {
		return err
	}
	return Unmarshal(data, v)
}
