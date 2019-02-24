package json

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

var (
	// Single-quote or double-quote quoted strings.
	stringPattern  = `(?:\'(?:\\.|[^\\\'])*\'|\"(?:\\.|[^\\\"])*\")`
	importRegexp   = regexp.MustCompile(`"@import\((.+)\)"`)
	commentsRegexp = regexp.MustCompile(`(?ms)` +
		// Inline comments begin with "//".
		`//(?U:.*)$` +
		// Paragraph comments within "/*" and "*/".
		`|/\*(?U:.*)\*/` +
		// Comment-like strings, which should be reserved.
		`|` + stringPattern,
	)
	trailingObjectCommasRegexp = regexp.MustCompile(`(?:,)\s*}` +
		// Literal strings which should be reserved.
		`|` + stringPattern,
	)
	trailingArrayCommasRegexp = regexp.MustCompile(`(?:,)\s*]` +
		// Literal strings which should be reserved.
		`|` + stringPattern,
	)
)

type Decoder interface {
	Decode(v interface{}) error
}

type Encoder interface {
	Encode(v interface{}) error
}

type extDecoder struct {
	reader io.Reader

	mu         sync.Mutex
	importRoot string
}

func NewExtDecoder(r io.Reader) *extDecoder {
	return &extDecoder{reader: r}
}

func (r *extDecoder) SetImportRoot(path string) *extDecoder {
	r.mu.Lock()
	r.importRoot = path
	r.mu.Unlock()
	return r
}

func (r *extDecoder) Decode(v interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

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

	data, err := replaceImports(content, r.importRoot)
	if err != nil {
		return err
	}
	data = removeComments(data)
	data = fixTrailingCommas(data)
	return Unmarshal(data, v)
}

func UnmarshalExt(data []byte, v interface{}) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	data, err = replaceImports(data, wd)
	if err != nil {
		return err
	}
	data = removeComments(data)
	data = fixTrailingCommas(data)
	return Unmarshal(data, v)
}

func removeComments(data []byte) []byte {
	return commentsRegexp.ReplaceAllFunc(data, func(src []byte) []byte {
		if src[0] == '/' {
			return []byte("")
		}
		return src
	})
}

func fixTrailingCommas(data []byte) []byte {
	// Fix objects {} first.
	data = trailingObjectCommasRegexp.ReplaceAllFunc(data, func(src []byte) []byte {
		if src[0] == ',' {
			return []byte("}")
		}
		return src
	})
	// Then fix arrays/lists [].
	data = trailingArrayCommasRegexp.ReplaceAllFunc(data, func(src []byte) []byte {
		if src[0] == ',' {
			return []byte("]")
		}
		return src
	})
	return data
}

func replaceImports(data []byte, importRoot string) ([]byte, error) {
	errs := make([]error, 0)
	data = importRegexp.ReplaceAllFunc(data, func(src []byte) []byte {
		subMatches := importRegexp.FindSubmatch(src)
		includedPath := filepath.Join(importRoot, string(subMatches[1]))
		included, err := ioutil.ReadFile(includedPath)
		if err != nil {
			errs = append(errs, err)
			return src
		}
		return included
	})

	if len(errs) != 0 {
		var errStrings = make([]string, len(errs))
		for i, e := range errs {
			errStrings[i] = e.Error()
		}
		return nil, errors.New(strings.Join(errStrings, "; "))
	}
	return data, nil
}
