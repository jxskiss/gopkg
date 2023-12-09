package yamlx

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Marshal serializes the value provided into a YAML document. The structure
// of the generated document will reflect the structure of the value itself.
// Maps and pointers (to struct, string, int, etc.) are accepted as the in value.
//
// See yaml.Marshal for detail docs.
func Marshal(in any) (out []byte, err error) {
	return yaml.Marshal(in)
}

// Unmarshal decodes the first document found within the in byte slice
// and assigns decoded values into the out value.
//
// Maps and pointers (to a struct, string, int, etc.) are accepted as out
// values. If an internal pointer within a struct is not initialized,
// the yaml package will initialize it if necessary for unmarshalling
// the provided data. The out parameter must not be nil.
//
// See yaml.Unmarshal for detail docs.
//
// Note that this package adds extra features on the standard YAML syntax,
// such as "reading environment variables", "file including",
// "reference using gjson JSON path expression",
// "reference using named variables", "function calling", etc.
func Unmarshal(in []byte, v any, options ...Option) error {
	par := newParser(in, options...)
	err := par.parse()
	if err != nil {
		return fmt.Errorf("cannot parse yaml data: %w", err)
	}
	err = par.Unmarshal(v)
	if err != nil {
		return fmt.Errorf("cannot unmarshal yaml data: %w", err)
	}
	return nil
}

// Load is a shortcut function to read YAML data directly from a file.
func Load(filename string, v any, options ...Option) error {
	yamlText, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("cannot read yaml file: %w", err)
	}
	absFilename, _ := filepath.Abs(filename)
	par := newParser(yamlText, options...)
	if absFilename != "" {
		par.filename = absFilename
		par.incStack = []string{absFilename}
	}
	err = par.parse()
	if err != nil {
		return fmt.Errorf("cannot parse yaml data: %w", err)
	}
	err = par.Unmarshal(v)
	if err != nil {
		return fmt.Errorf("cannot unmarshal yaml data: %w", err)
	}
	return nil
}
