package validat

import (
	"context"
	"fmt"
	"strconv"

	"github.com/jxskiss/gopkg/v2/internal/constraints"
	"github.com/jxskiss/gopkg/v2/unsafe/reflectx"
)

func GreaterThanZero[T constraints.Integer](name string, value T) RuleFunc {
	return func(ctx context.Context, result *Result) error {
		if value <= 0 {
			return fmt.Errorf("%s value %v <= 0", name, value)
		}
		return nil
	}
}

func Int64GreaterThanZero[T Int64OrString](name string, value T) RuleFunc {
	return func(ctx context.Context, result *Result) error {
		intVal, err := parseInt64(value)
		if err != nil {
			return err
		}
		if intVal <= 0 {
			return fmt.Errorf("%s value %v <= 0", name, value)
		}
		if name != "" {
			result.Extra.Set(name, intVal)
		}
		return nil
	}
}

func LessThanOrEqual[T constraints.Integer](name string, limit, value T) RuleFunc {
	return func(ctx context.Context, result *Result) error {
		if value > limit {
			return fmt.Errorf("%s value %d > %d", name, value, limit)
		}
		return nil
	}
}

func StringIntegerGreaterThanZero(name string, value string) RuleFunc {
	return func(ctx context.Context, result *Result) error {
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("%s value is not integer: %v", name, value)
		}
		if intVal <= 0 {
			return fmt.Errorf("%s value %v <= 0", name, value)
		}
		if name != "" {
			result.Extra.Set(name, intVal)
		}
		return nil
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
	return func(ctx context.Context, result *Result) error {
		if value >= min && value <= max {
			return nil
		}
		return fmt.Errorf("%s value %v is not in range [%v, %v]", name, value, min, max)
	}
}

func InRangeMode[T constraints.Integer](name string, mode RangeMode, min, max T, value T) RuleFunc {
	return func(ctx context.Context, result *Result) error {
		switch mode {
		case GtAndLte:
			if value > min && value <= max {
				return nil
			}
			return fmt.Errorf("%s value %v is not in range (%v, %v]", name, value, min, max)
		case GtAndLt:
			if value > min && value < max {
				return nil
			}
			return fmt.Errorf("%s value %v is not in range (%v, %v)", name, value, min, max)
		case GteAndLte:
			if value >= min && value <= max {
				return nil
			}
			return fmt.Errorf("%s value %v is not in range [%v, %v]", name, value, min, max)
		case GteAndLt:
			if value >= min && value < max {
				return nil
			}
			return fmt.Errorf("%s value %v is not in range [%v, %v)", name, value, min, max)
		default:
			return fmt.Errorf("%s unknown range mode %v", name, mode)
		}
	}
}

func ParseStringsToInt64s(name string, values []string) RuleFunc {
	return func(ctx context.Context, result *Result) error {
		out := make([]int64, 0, len(values))
		for _, v := range values {
			intVal, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return fmt.Errorf("%s value %s is not integer", name, v)
			}
			out = append(out, intVal)
		}
		if name != "" {
			result.Extra.Set(name, out)
		}
		return nil
	}
}

func ParseStringsToInt64Map(name string, values []string) RuleFunc {
	return func(ctx context.Context, result *Result) error {
		out := make(map[int64]bool, len(values))
		for _, v := range values {
			intVal, err := strconv.ParseInt(v, 10, 64)
			if err != nil {
				return fmt.Errorf("%s value %s is not integer", name, v)
			}
			out[intVal] = true
		}
		if name != "" {
			result.Extra.Set(name, out)
		}
		return nil
	}
}

func NotNil(name string, value any) RuleFunc {
	return func(ctx context.Context, result *Result) error {
		if reflectx.IsNil(value) {
			return fmt.Errorf("%s value is nil", name)
		}
		return nil
	}
}
