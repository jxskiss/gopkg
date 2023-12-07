package vdutil

import (
	"context"
	"fmt"
	"regexp"

	"github.com/jxskiss/gopkg/v2/perf/lru"
)

var reCache lru.Interface[string, *regexp.Regexp]

func EnableRegexpCache(cache lru.Interface[string, *regexp.Regexp]) {
	reCache = cache
}

type RegexpOrString interface {
	*regexp.Regexp | string
}

func MatchRegexp[T RegexpOrString](name string, pattern T, value string) RuleFunc {
	re, isRegexp := any(pattern).(*regexp.Regexp)
	if isRegexp {
		return func(ctx context.Context, result *Result) (any, error) {
			var err error
			match := re.MatchString(value)
			if !match {
				err = &ValidationError{Name: name, Err: fmt.Errorf("value %q not match regexp", value)}
			}
			return value, err
		}
	}

	return func(ctx context.Context, result *Result) (any, error) {
		var err error
		reStr := any(pattern).(string)
		if reCache != nil {
			re, _, _ = reCache.Get(reStr)
		}
		if re == nil {
			re, err = regexp.Compile(reStr)
			if err != nil {
				return value, fmt.Errorf("cannot compile regexp %q: %w", reStr, err)
			}
			if reCache != nil {
				reCache.Set(reStr, re, 0)
			}
		}
		match := re.MatchString(value)
		if !match {
			err = &ValidationError{Name: name, Err: fmt.Errorf("value %q not match regexp", value)}
		}
		return value, err
	}
}
