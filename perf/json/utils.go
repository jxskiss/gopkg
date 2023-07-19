package json

import (
	"io"
	"os"
)

// Load reads JSON-encoded data from the named file at path and stores
// the result in the value pointed to by v.
func Load(path string, v any) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	err = NewDecoder(file).Decode(v)
	return err
}

// Dump writes v to the named file at path using JSON encoding.
// It disables HTMLEscape.
// Optionally indent can be applied to the output,
// empty prefix and indent disables indentation.
// The output is friendly to read by humans.
func Dump(path string, v any, prefix, indent string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return Fdump(file, v, prefix, indent)
}

// Fdump writes v to the given io.Writer using JSON encoding.
// It disables HTMLEscape.
// Optionally indent can be applied to the output,
// empty prefix and indent disables indentation.
// The output is friendly to read by humans.
func Fdump(w io.Writer, v any, prefix, indent string) error {
	return NewEncoder(w).
		SetEscapeHTML(false).
		SetIndent(prefix, indent).
		Encode(v)
}
