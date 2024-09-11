package vdutil

import (
	"context"
	"errors"

	"github.com/jxskiss/gopkg/v2/easy/ezmap"
)

// Rule represents a validating rule.
type Rule interface {
	Validate(ctx context.Context, result *Result) (any, error)
}

// RuleFunc implements the interface Rule.
type RuleFunc func(ctx context.Context, result *Result) (any, error)

func (f RuleFunc) Validate(ctx context.Context, result *Result) (any, error) {
	return f(ctx, result)
}

// Result can be used to pass data from validating rules to the caller.
type Result struct {
	// Data allows validating rules to pass data to the caller.
	Data ezmap.Map

	// Validating rules can optionally add detail information to ErrDetails
	// to inform the caller.
	ErrDetails []any

	// IsValidationError tells whether the error returned by Validate
	// is a ValidationError or not.
	IsValidationError bool
}

// ValidationError indicates that the data is invalid.
// Name and Err tells which data and underlying error returned by a Rule.
// Note, a Rule should not return this error on programming error
// or internal system error.
type ValidationError struct {
	Name string
	Err  error
}

func (e *ValidationError) Error() string {
	if e.Name == "" {
		return e.Err.Error()
	}
	return e.Name + ": " + e.Err.Error()
}

func (e *ValidationError) Unwrap() error { return e.Err }

// Validate runs validating rules, it returns a Result and the first error
// returned from the rules. If a rule returns an error, it returns
// and the remaining rules are not executed.
func Validate(ctx context.Context, rules ...Rule) (result *Result, err error) {
	for _, r := range rules {
		if rr, ok := r.(*useResult); ok {
			result = (*Result)(rr)
		}
	}
	if result == nil {
		result = &Result{}
	}
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

// UseResult tells validating rules to use the given result.
func UseResult(result *Result) Rule {
	return (*useResult)(result)
}

type useResult Result

func (*useResult) Validate(_ context.Context, _ *Result) (any, error) {
	return nil, nil
}
