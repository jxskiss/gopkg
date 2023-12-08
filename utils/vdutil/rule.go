package vdutil

import (
	"context"
	"errors"

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

	IsValidationError bool
}

type ValidationError struct {
	Name string
	Err  error
}

func (e *ValidationError) Error() string { return e.Name + ": " + e.Err.Error() }

func (e *ValidationError) Unwrap() error { return e.Err }

func Validate(ctx context.Context, rules ...Rule) (result *Result, err error) {
	result = &Result{}
	for _, rule := range rules {
		_, err = rule.Validate(ctx, result)
		if err != nil {
			break
		}
	}
	if err != nil {
		result.IsValidationError = errors.As(err, new(*ValidationError))
	}
	return result, err
}
