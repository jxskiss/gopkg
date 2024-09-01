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

// GreaterThanZero validates that value is greater than zero.
// value can be either an integer or a string.
// If save is true, the result integer value will be saved to
// Result.Data using name as key.
func GreaterThanZero[T IntegerOrString](name string, value T, save bool) RuleFunc {
	return _greaterThanZero(name, value, save)
}

// AllElementsGreaterThanZero validates that all elements in slice
// are greater than zero.
// If save is true, the result integer slice will be saved to
// Result.Data using name as key.
func AllElementsGreaterThanZero[T IntegerOrString](name string, slice []T, save bool) RuleFunc {
	var zero T
	typ := reflect.TypeOf(zero)
	switch typ.Kind() {
	case reflect.String:
		return func(_ context.Context, result *Result) (any, error) {
			out := make([]int64, 0, len(slice))
			for _, elem := range slice {
				i64Val, err := strconv.ParseInt(reflect.ValueOf(elem).String(), 10, 64)
				if err != nil {
					return nil, &ValidationError{Name: name, Err: fmt.Errorf("slice has non-integer element")}
				}
				if i64Val <= 0 {
					return nil, &ValidationError{Name: name, Err: fmt.Errorf("slice element %v <= 0", elem)}
				}
			}
			if save && name != "" {
				result.Data.Set(name, out)
			}
			return out, nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return func(_ context.Context, result *Result) (any, error) {
			for _, elem := range slice {
				i64Val := reflect.ValueOf(elem).Int()
				if i64Val <= 0 {
					return nil, &ValidationError{Name: name, Err: fmt.Errorf("slice element %v <= 0", elem)}
				}
			}
			if save && name != "" {
				result.Data.Set(name, slice)
			}
			return slice, nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return func(_ context.Context, result *Result) (any, error) {
			for _, elem := range slice {
				u64Val := reflect.ValueOf(elem).Uint()
				if u64Val <= 0 {
					return nil, &ValidationError{Name: name, Err: fmt.Errorf("slice element %v <= 0", elem)}
				}
			}
			if save && name != "" {
				result.Data.Set(name, slice)
			}
			return slice, nil
		}
	default:
		panic("bug: unreachable code")
	}
}

// AllElementsNotZero validates that all elements in slice
// are not equal to the zero value of type T.
func AllElementsNotZero[T comparable](name string, slice []T) RuleFunc {
	return func(_ context.Context, _ *Result) (any, error) {
		var zero T
		for _, elem := range slice {
			if elem == zero {
				return nil, &ValidationError{Name: name, Err: fmt.Errorf("slice has zero element")}
			}
		}
		return nil, nil
	}
}

func _greaterThanZero(name string, value any, save bool) RuleFunc {
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.String:
		return func(_ context.Context, result *Result) (any, error) {
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
		return func(_ context.Context, result *Result) (any, error) {
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
		return func(_ context.Context, result *Result) (any, error) {
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

// LessThanOrEqual validates that value <= limit.
func LessThanOrEqual[T constraints.Ordered](name string, limit, value T) RuleFunc {
	return func(_ context.Context, _ *Result) (any, error) {
		var err error
		if value > limit {
			err = &ValidationError{Name: name, Err: fmt.Errorf("value %v > %v", value, limit)}
		}
		return value, err
	}
}

// RangeMode tells InRangeMode how to handle lower and upper bound
// when validating a value against a range.
type RangeMode int

const (
	GtAndLte  RangeMode = iota // min < x && x <= max
	GtAndLt                    // min < x && x < max
	GteAndLte                  // min <= x && x <= max
	GteAndLt                   // min <= x && x < max
)

// InRange validates that value >= min and value <= max.
func InRange[T constraints.Ordered](name string, min, max T, value T) RuleFunc {
	return InRangeMode(name, GteAndLte, min, max, value)
}

// InRangeMode validates that value is in range of min and max,
// according to RangeMode.
func InRangeMode[T constraints.Ordered](name string, mode RangeMode, min, max T, value T) RuleFunc {
	return func(_ context.Context, _ *Result) (any, error) {
		var err error
		switch mode {
		case GtAndLte:
			if !(value > min && value <= max) {
				err = &ValidationError{Name: name, Err: fmt.Errorf("value %v is not in range (%v, %v]", value, min, max)}
			}
		case GtAndLt:
			if !(value > min && value < max) {
				err = &ValidationError{Name: name, Err: fmt.Errorf("value %v is not in range (%v, %v)", value, min, max)}
			}
		case GteAndLte:
			if !(value >= min && value <= max) {
				err = &ValidationError{Name: name, Err: fmt.Errorf("value %v is not in range [%v, %v]", value, min, max)}
			}
		case GteAndLt:
			if !(value >= min && value < max) {
				err = &ValidationError{Name: name, Err: fmt.Errorf("value %v is not in range [%v, %v)", value, min, max)}
			}
		default:
			err = fmt.Errorf("%s: unknown range mode %v", name, mode)
		}
		return value, err
	}
}

// ParseStrsToInt64Slice validates all elements in values are valid integer
// and convert values to be an []int64 slice, the result slice will be
// saved to Result.Data using name as key.
func ParseStrsToInt64Slice[T ~string](name string, values []T) RuleFunc {
	return func(_ context.Context, result *Result) (any, error) {
		out := make([]int64, 0, len(values))
		for _, v := range values {
			intVal, err := strconv.ParseInt(string(v), 10, 64)
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

// ParseStrsToInt64Map validates all elements in values are valid integer
// and convert values to be a map[int64]bool, the result map will be
// saved to Result.Data using name as key.
func ParseStrsToInt64Map[T ~string](name string, values []T) RuleFunc {
	return func(_ context.Context, result *Result) (any, error) {
		out := make(map[int64]bool, len(values))
		for _, v := range values {
			intVal, err := strconv.ParseInt(string(v), 10, 64)
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

// NotNil validates value is not nil (e.g. nil pointer, nil slice, nil map).
func NotNil(name string, value any) RuleFunc {
	return func(_ context.Context, _ *Result) (any, error) {
		var err error
		if reflectx.IsNil(value) {
			err = &ValidationError{Name: name, Err: errors.New("value is nil")}
		}
		return value, err
	}
}
