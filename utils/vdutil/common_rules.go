package vdutil

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/jxskiss/gopkg/v2/internal/constraints"
	"github.com/jxskiss/gopkg/v2/unsafe/reflectx"
)

type IntegerOrString interface {
	constraints.Integer | ~string
}

func GreaterThanZero[T IntegerOrString](name string, value T, save bool) RuleFunc {
	return _greaterThanZero(name, value, save)
}

func _greaterThanZero(name string, value any, save bool) RuleFunc {
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.String:
		return func(ctx context.Context, result *Result) (any, error) {
			i64Val, err := strconv.ParseInt(rv.String(), 10, 64)
			if err != nil {
				return int64(0), &ValidationError{Name: name, Err: fmt.Errorf("value %v is not integer: %w", value, err)}
			}
			if i64Val <= 0 {
				return i64Val, &ValidationError{Name: name, Err: fmt.Errorf("value %v <= 0", value)}
			}
			if save && name != "" {
				result.Data.Set(name, i64Val)
			}
			return i64Val, nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return func(ctx context.Context, result *Result) (any, error) {
			i64Val := rv.Int()
			if i64Val <= 0 {
				return value, &ValidationError{Name: name, Err: fmt.Errorf("value %v <= 0", value)}
			}
			if save && name != "" {
				result.Data.Set(name, value)
			}
			return value, nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return func(ctx context.Context, result *Result) (any, error) {
			u64Val := rv.Uint()
			if u64Val <= 0 {
				return value, &ValidationError{Name: name, Err: fmt.Errorf("value %v <= 0", value)}
			}
			if save && name != "" {
				result.Data.Set(name, value)
			}
			return value, nil
		}
	default:
		panic("bug: unreachable code")
	}
}

func LessThanOrEqual[T constraints.Integer](name string, limit, value T) RuleFunc {
	return func(ctx context.Context, result *Result) (any, error) {
		var err error
		if value > limit {
			err = &ValidationError{Name: name, Err: fmt.Errorf("value %d > %d", value, limit)}
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
			err = &ValidationError{Name: name, Err: fmt.Errorf("value %d is not in range [%d, %d]", value, min, max)}
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
				err = &ValidationError{Name: name, Err: fmt.Errorf("value %v is not in range (%d, %d]", value, min, max)}
			}
		case GtAndLt:
			if !(value > min && value < max) {
				err = &ValidationError{Name: name, Err: fmt.Errorf("value %d is not in range (%d, %d)", value, min, max)}
			}
		case GteAndLte:
			if !(value >= min && value <= max) {
				err = &ValidationError{Name: name, Err: fmt.Errorf("value %d is not in range [%d, %d]", value, min, max)}
			}
		case GteAndLt:
			if !(value >= min && value < max) {
				err = &ValidationError{Name: name, Err: fmt.Errorf("value %d is not in range [%d, %d)", value, min, max)}
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
				return nil, &ValidationError{Name: name, Err: fmt.Errorf("value %v is not integer", v)}
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
				return nil, &ValidationError{Name: name, Err: fmt.Errorf("value %v is not integer", v)}
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
			err = &ValidationError{Name: name, Err: errors.New("value is nil")}
		}
		return value, err
	}
}
