package ptr

import (
	"reflect"
	"strconv"
	"time"
	"unsafe"

	"github.com/jxskiss/gopkg/v2/internal/constraints"
)

func Bool(v bool) *bool                       { return &v }
func Time(v time.Time) *time.Time             { return &v }
func Duration(v time.Duration) *time.Duration { return &v }

func String[T ~string | constraints.Integer](v T) *string {
	if s, ok := any(v).(string); ok {
		return &s
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.String:
		s := rv.String()
		return &s
	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Int:
		s := strconv.FormatInt(rv.Int(), 10)
		return &s
	case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uint, reflect.Uintptr:
		s := strconv.FormatUint(rv.Uint(), 10)
		return &s
	}
	panic("bug: unreachable code")
}

func Int[T constraints.Integer](v T) *int {
	x := int(v)
	return &x
}

func Int8[T constraints.Integer](v T) *int8 {
	x := int8(v)
	return &x
}

func Int16[T constraints.Integer](v T) *int16 {
	x := int16(v)
	return &x
}

func Int32[T constraints.Integer](v T) *int32 {
	x := int32(v)
	return &x
}

func Int64[T constraints.Integer](v T) *int64 {
	x := int64(v)
	return &x
}

func Uint[T constraints.Integer](v T) *uint {
	x := uint(v)
	return &x
}

func Uint8[T constraints.Integer](v T) *uint8 {
	x := uint8(v)
	return &x
}

func Uint16[T constraints.Integer](v T) *uint16 {
	x := uint16(v)
	return &x
}

func Uint32[T constraints.Integer](v T) *uint32 {
	x := uint32(v)
	return &x
}

func Uint64[T constraints.Integer](v T) *uint64 {
	x := uint64(v)
	return &x
}

func Float32[T constraints.RealNumber](v T) *float32 {
	x := float32(v)
	return &x
}

func Float64[T constraints.RealNumber](v T) *float64 {
	x := float64(v)
	return &x
}

func DerefBool(v *bool) bool                       { return Deref(v) }
func DerefTime(v *time.Time) time.Time             { return Deref(v) }
func DerefDuration(v *time.Duration) time.Duration { return Deref(v) }

func DerefString[T ~string](v *T) string {
	return Deref((*string)(unsafe.Pointer(v)))
}

func DerefInt[T constraints.Integer](v *T) (ret int) {
	if v != nil {
		ret = int(*v)
	}
	return
}

func DerefInt8[T constraints.Integer](v *T) (ret int8) {
	if v != nil {
		ret = int8(*v)
	}
	return
}

func DerefInt16[T constraints.Integer](v *T) (ret int16) {
	if v != nil {
		ret = int16(*v)
	}
	return
}

func DerefInt32[T constraints.Integer](v *T) (ret int32) {
	if v != nil {
		ret = int32(*v)
	}
	return
}

func DerefInt64[T constraints.Integer](v *T) (ret int64) {
	if v != nil {
		ret = int64(*v)
	}
	return
}

func DerefUint[T constraints.Integer](v *T) (ret uint) {
	if v != nil {
		ret = uint(*v)
	}
	return
}

func DerefUint8[T constraints.Integer](v *T) (ret uint8) {
	if v != nil {
		ret = uint8(*v)
	}
	return
}

func DerefUint16[T constraints.Integer](v *T) (ret uint16) {
	if v != nil {
		ret = uint16(*v)
	}
	return
}

func DerefUint32[T constraints.Integer](v *T) (ret uint32) {
	if v != nil {
		ret = uint32(*v)
	}
	return
}

func DerefUint64[T constraints.Integer](v *T) (ret uint64) {
	if v != nil {
		ret = uint64(*v)
	}
	return
}

func DerefFloat32[T constraints.RealNumber](v *T) (ret float32) {
	if v != nil {
		ret = float32(*v)
	}
	return
}

func DerefFloat64[T constraints.RealNumber](v *T) (ret float64) {
	if v != nil {
		ret = float64(*v)
	}
	return
}
