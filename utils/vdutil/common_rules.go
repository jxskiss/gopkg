package validat

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/jxskiss/gopkg/v2/internal/constraints"
	"github.com/jxskiss/gopkg/v2/unsafe/reflectx"
)

func GreaterThanZero[T constraints.Integer](name string, value T) RuleFunc {
	return func(ctx context.Context, result *Result) (any, error) {
		var err error
		if value <= 0 {
			err = &ValidatingError{Name: name, Err: fmt.Errorf("value %v <= 0", value)}
		}
		return value, err
	}
}

func Int64GreaterThanZero[T Int64OrString](name string, value T, save bool) RuleFunc {
	var zero int64
	return func(ctx context.Context, result *Result) (any, error) {
		intVal, err := parseInt64(value)
		if err != nil {
			return zero, &ValidatingError{Name: name, Err: fmt.Errorf("value %v is not integer: %w", value, err)}
		}
		if intVal <= 0 {
			return zero, &ValidatingError{Name: name, Err: fmt.Errorf("value %v <= 0", value)}
		}
		if save && name != "" {
			result.Data.Set(name, intVal)
		}
		return intVal, nil
	}
}

func LessThanOrEqual[T constraints.Integer](name string, limit, value T) RuleFunc {
	return func(ctx context.Context, result *Result) (any, error) {
		var err error
		if value > limit {
			err = &ValidatingError{Name: name, Err: fmt.Errorf("value %d > %d", value, limit)}
		}
		return value, err
	}
}

type RangeMode int

const (
	GtAndLte  RangeMode = iota // min < x && x <= max
	GtAndLt                    // min < x && x < max
	GteAndLte                  // min <= x && x <= max
	GteAndLt                   // min <= x && x < max
)

func InRange[T constraints.Integer](name string, min, max T, value T) RuleFunc {
	return func(ctx context.Context, result *Result) (any, error) {
		var err error
		if value < min || value > max {
			err = &ValidatingError{Name: name, Err: fmt.Errorf("value %d is not in range [%d, %d]", value, min, max)}
		}
		return value, err
	}
}

func InRangeMode[T constraints.Integer](name string, mode RangeMode, min, max T, value T) RuleFunc {
	return func(ctx context.Context, result *Result) (any, error) {
		var err error
		switch mode {
		case GtAndLte:
			if !(value > min && value <= max) {
				err = &ValidatingError{Name: name, Err: fmt.Errorf("value %v is not in range (%d, %d]", value, min, max)}
			}
		case GtAndLt:
			if !(value > min && value < max) {
				err = &ValidatingError{Name: name, Err: fmt.Errorf("value %d is not in range (%d, %d)", value, min, max)}
			}
		case GteAndLte:
			if !(value >= min && value <= max) {
				err = &ValidatingError{Name: name, Err: fmt.Errorf("value %d is not in range [%d, %d]", value, min, max)}
			}
		case GteAndLt:
			if !(value >= min && value < max) {
				err = &ValidatingError{Name: name, Err: fmt.Errorf("value %d is not in range [%d, %d)", value, min, max)}
			}
		default:
			err = fmt.Errorf("%s: unknown range mode %v", name, mode)
		}
		return value, err
	}
}

func ParseStrsToInt64Slice(name string, values []string) RuleFunc {
	return func(ctx context.Context, result *Result) (any, error) {
		out := make([]int64, 0, len(values))
		for _, v := range values {
			intVal, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return nil, &ValidatingError{Name: name, Err: fmt.Errorf("value %v is not integer", v)}
			}
			out = append(out, intVal)
		}
		if name != "" {
			result.Data.Set(name, out)
		}
		return out, nil
	}
}

func ParseStrsToInt64Map(name string, values []string) RuleFunc {
	return func(ctx context.Context, result *Result) (any, error) {
		out := make(map[int64]bool, len(values))
		for _, v := range values {
			intVal, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return nil, &ValidatingError{Name: name, Err: fmt.Errorf("value %v is not integer", v)}
			}
			out[intVal] = true
		}
		if name != "" {
			result.Data.Set(name, out)
		}
		return out, nil
	}
}

func NotNil(name string, value any) RuleFunc {
	return func(ctx context.Context, result *Result) (any, error) {
		var err error
		if reflectx.IsNil(value) {
			err = &ValidatingError{Name: name, Err: errors.New("value is nil")}
		}
		return value, err
	}
}
