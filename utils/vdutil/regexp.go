package vdutil

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	"github.com/jxskiss/gopkg/v2/perf/lru"
)

var reCache lru.Interface[string, *regexp.Regexp]

// EnableRegexpCache sets an LRU cache to cache compiled regular expressions.
func EnableRegexpCache(cache lru.Interface[string, *regexp.Regexp]) {
	reCache = cache
}

type RegexpOrString interface {
	*regexp.Regexp | string
}

// MatchRegexp validates value match the regular expression pattern.
// pattern can be either a string or a compiled *regexp.Regexp.
//
// If pattern is a string and cache is enabled by calling EnableRegexpCache,
// the compiled regular expression will be cached for reuse.
func MatchRegexp[T RegexpOrString](name string, pattern T, value string) RuleFunc {
	re, compileErr := compileRegexp(pattern)
	return func(_ context.Context, _ *Result) (any, error) {
		if compileErr != nil {
			return value, compileErr
		}
		match := re.MatchString(value)
		if !match {
			return value, &ValidationError{Name: name, Err: errors.New("value not match regexp")}
		}
		return value, nil
	}
}

func compileRegexp(expr any) (*regexp.Regexp, error) {
	if re, ok := expr.(*regexp.Regexp); ok {
		return re, nil
	}
	var re *regexp.Regexp
	exprStr := expr.(string)
	if reCache != nil {
		re, _, _ = reCache.Get(exprStr)
	}
	if re == nil {
		var err error
		re, err = regexp.Compile(exprStr)
		if err != nil {
			return nil, fmt.Errorf("cannot compile regexp %q: %w", exprStr, err)
		}
		if reCache != nil {
			reCache.Set(exprStr, re, 0)
		}
	}
	return re, nil
}
