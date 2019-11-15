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

	maxImportDepth = 10
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

	data, err := fixJSON(content, r.importRoot, 1)
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
	data, err = fixJSON(data, wd, 1)
	if err != nil {
		return err
	}
	return Unmarshal(data, v)
}

func fixJSON(data []byte, importRoot string, depth int) ([]byte, error) {
	if depth > maxImportDepth {
		return nil, errors.New("max depth exceeded")
	}
	data = removeComments(data)
	data = fixTrailingCommas(data)
	data, replaced, err := replaceImports(data, importRoot)
	if err != nil {
		return nil, err
	}
	if replaced {
		return fixJSON(data, importRoot, depth+1)
	}
	return data, nil
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

func replaceImports(data []byte, importRoot string) (result []byte, replaced bool, err error) {
	replaceErrs := make([]error, 0)
	result = importRegexp.ReplaceAllFunc(data, func(src []byte) []byte {
		replaced = true
		subMatches := importRegexp.FindSubmatch(src)
		includedPath := filepath.Join(importRoot, string(subMatches[1]))
		included, err := ioutil.ReadFile(includedPath)
		if err != nil {
			replaceErrs = append(replaceErrs, err)
			return src
		}
		return included
	})

	if len(replaceErrs) != 0 {
		var errStrings = make([]string, len(replaceErrs))
		for i, e := range replaceErrs {
			errStrings[i] = e.Error()
		}
		err = errors.New(strings.Join(errStrings, "; "))
	}
	return
}
