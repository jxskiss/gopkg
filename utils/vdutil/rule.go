package validat

import (
	"context"

	"github.com/jxskiss/gopkg/v2/easy/ezmap"
)

type Rule interface {
	Validate(ctx context.Context, result *Result) error
}

type RuleFunc func(ctx context.Context, result *Result) error

func (f RuleFunc) Validate(ctx context.Context, result *Result) error {
	return f(ctx, result)
}

type Result struct {
	IsInternalError bool

	Extra ezmap.Map
}

func Validate(ctx context.Context, rules ...Rule) (*Result, error) {
	ret := &Result{}
	for _, rule := range rules {
		err := rule.Validate(ctx, ret)
		if err != nil {
			return ret, err
		}
	}
	return ret, nil
}
