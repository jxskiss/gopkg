package validat

import (
	"context"

	"github.com/jxskiss/gopkg/v2/easy/ezmap"
)

type Rule interface {
	Validate(ctx context.Context, result *Result) (any, error)
}

type RuleFunc func(ctx context.Context, result *Result) (any, error)

func (f RuleFunc) Validate(ctx context.Context, result *Result) (any, error) {
	return f(ctx, result)
}

type Result struct {
	Data       ezmap.Map
	ErrDetails []any
}

type ValidatingError struct {
	Name string
	Err  error
}

func (e *ValidatingError) Error() string { return e.Name + ": " + e.Err.Error() }

func (e *ValidatingError) Unwrap() error { return e.Err }

func Validate(ctx context.Context, rules ...Rule) (*Result, error) {
	ret := &Result{}
	for _, rule := range rules {
		_, err := rule.Validate(ctx, ret)
		if err != nil {
			return ret, err
		}
	}
	return ret, nil
}
