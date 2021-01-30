package easy

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	htmltemplate "html/template"
	texttemplate "text/template"
)

// ParseHTMLTemplates parses files under `rootDir` which matches the regular
// expression `rePattern`. Optionally a `funcMap` can be specified to use
// with the parsed templates.
//
// The returned Template holds the parsed templates under the root directory,
// template can be retrieved using Template.Lookup(name), where name is the
// file path relative to rootDir, without leading "./".
func ParseHTMLTemplates(rootDir string, rePattern string, funcMap htmltemplate.FuncMap) (*htmltemplate.Template, error) {
	t := htmltemplate.New("").Funcs(funcMap)
	err := parseTemplates(rootDir, rePattern, func(name string, text []byte) error {
		_, e1 := t.New(name).Parse(string(text))
		return e1
	})
	if err != nil {
		return nil, err
	}
	return t, nil
}

// ParseTextTemplates parses files under `rootDir` which matches the regular
// expression `rePattern`. Optionally a `funcMap` can be specified to use
// with the parsed templates.
//
// The returned Template holds the parsed templates under the root directory,
// template can be retrieved using Template.Lookup(name), where name is the
// file path relative to rootDir, without leading "./".
func ParseTextTemplates(rootDir string, rePattern string, funcMap texttemplate.FuncMap) (*texttemplate.Template, error) {
	t := texttemplate.New("").Funcs(funcMap)
	err := parseTemplates(rootDir, rePattern, func(name string, text []byte) error {
		_, e1 := t.New(name).Parse(string(text))
		return e1
	})
	if err != nil {
		return nil, err
	}
	return t, nil
}

// https://stackoverflow.com/a/50581032
func parseTemplates(rootDir string, rePattern string, add func(name string, text []byte) error) error {
	cleanRoot := filepath.Clean(rootDir)
	pfx := len(cleanRoot) + 1
	re, err := regexp.Compile(rePattern)
	if err != nil {
		return err
	}
	return filepath.Walk(cleanRoot, func(path string, info os.FileInfo, e1 error) error {
		if e1 != nil {
			return e1
		}
		if info.IsDir() {
			return nil
		}
		name := path[pfx:]
		if !re.MatchString(name) {
			return nil
		}
		text, e2 := ioutil.ReadFile(path)
		if e2 != nil {
			return e2
		}
		e2 = add(name, text)
		return e2
	})
}
