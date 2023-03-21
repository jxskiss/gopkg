package ptr

import (
	"reflect"
	"strconv"

	"github.com/jxskiss/gopkg/v2/internal/constraints"
)

// IntToStringp converts Integer x to a string pointer.
func IntToStringp[T constraints.Integer](x T) *string {
	rv := reflect.ValueOf(x)
	switch rv.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		s := strconv.FormatInt(rv.Int(), 10)
		return &s
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint, reflect.Uintptr:
		s := strconv.FormatUint(rv.Uint(), 10)
		return &s
	}
	panic("bug: unreachable code")
}

// IntpToStringp converts x to a string pointer.
// It returns nil if x is nil.
func IntpToStringp[T constraints.Integer](x *T) *string {
	if x == nil {
		return nil
	}
	rv := reflect.ValueOf(*x)
	switch rv.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		s := strconv.FormatInt(rv.Int(), 10)
		return &s
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint, reflect.Uintptr:
		s := strconv.FormatUint(rv.Uint(), 10)
		return &s
	}
	panic("bug: unreachable code")
}

// IntpToString converts x to a string.
// It returns an empty string if x is nil.
func IntpToString[T constraints.Integer](x *T) string {
	if x == nil {
		return ""
	}
	rv := reflect.ValueOf(*x)
	switch rv.Kind() {
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		s := strconv.FormatInt(rv.Int(), 10)
		return s
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint, reflect.Uintptr:
		s := strconv.FormatUint(rv.Uint(), 10)
		return s
	}
	panic("bug: unreachable code")
}
