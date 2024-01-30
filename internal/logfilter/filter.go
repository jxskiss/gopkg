package logfilter

import (
	"fmt"
	"strings"

	"github.com/gobwas/glob"
)

type FileNameFilter struct {
	rule      string
	allowRule string
	denyRule  string
	errs      []error

	allowAll   bool
	allowGlobs []glob.Glob

	denyAll   bool
	denyGlobs []glob.Glob
}

func NewFileNameFilter(rule string) (*FileNameFilter, []error) {
	lf := &FileNameFilter{rule: rule}
	if rule == "" {
		lf.allowAll = true
		return lf, nil
	}
	directives := strings.Split(rule, ";")
	for _, r := range directives {
		if r == "" {
			continue
		}
		if strings.HasPrefix(r, "allow=") {
			lf.parseAllowRule(r)
		}
		if strings.HasPrefix(r, "deny=") {
			lf.parseDenyRule(r)
		}
	}
	if len(lf.allowGlobs) == 0 {
		lf.allowAll = true
	}
	return lf, lf.errs
}

func (f *FileNameFilter) parseAllowRule(rule string) {
	f.allowRule = rule
	// Remove the prefix "allow=".
	globStrs := strings.Split(rule[6:], ",")
	for _, s := range globStrs {
		if s == "" {
			continue
		}
		if s == "all" {
			f.allowGlobs = nil
			break
		}
		g, err := glob.Compile("**"+s, '/')
		if err != nil {
			f.errs = append(f.errs, fmt.Errorf("cannot compile filter pattern %q: %w", s, err))
			continue
		}
		f.allowGlobs = append(f.allowGlobs, g)
	}
}

func (f *FileNameFilter) parseDenyRule(rule string) {
	f.denyRule = rule
	// Remove the prefix "deny=".
	globStrs := strings.Split(rule[5:], ",")
	for _, s := range globStrs {
		if s == "" {
			continue
		}
		if s == "all" {
			f.denyAll = true
			f.denyGlobs = nil
			break
		}
		g, err := glob.Compile("**"+s, '/')
		if err != nil {
			f.errs = append(f.errs, fmt.Errorf("cannot compile filter pattern %q: %w", s, err))
			continue
		}
		f.denyGlobs = append(f.denyGlobs, g)
	}
}

func (f *FileNameFilter) Allow(fileName string) bool {
	// not allow-all, check allow rules
	if !f.allowAll {
		for _, g := range f.allowGlobs {
			if g.Match(fileName) {
				return true
			}
		}
		return false
	}

	// allow-all, check deny rules
	if !f.denyAll {
		for _, g := range f.denyGlobs {
			if g.Match(fileName) {
				return false
			}
		}
	}
	return !f.denyAll
}
