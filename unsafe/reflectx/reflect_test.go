package reflectx

import (
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type dummyError struct{}

func (_ *dummyError) Error() string { return "dummyError" }

func TestIsNil(t *testing.T) {
	testcases := []struct {
		v    any
		want bool
	}{
		{nil, true},
		{(map[string]int)(nil), true},
		{([]string)(nil), true},
		{(*int)(nil), true},
		{(*simple)(nil), true},
		{error((*dummyError)(nil)), true},
		{map[string]int{}, false},
		{[]string{}, false},
		{1, false},
		{"abc", false},
		{simple{}, false},
		{&simple{}, false},
	}
	for i, tc := range testcases {
		got := IsNil(tc.v)
		assert.Equalf(t, tc.want, got, "i= %v, v = %q", i, tc.v)
	}
}

func TestCastInt(t *testing.T) {
	values := []any{
		int8(1), int8(math.MinInt8), int8(math.MaxInt8),
		int16(1), int16(math.MinInt16), int16(math.MaxInt16),
		int32(1), int32(math.MinInt32), int32(math.MaxInt32),
		int64(1), int64(math.MinInt64), int64(math.MaxInt64),
		uint8(0), uint8(1), uint8(math.MaxUint8),
		uint16(0), uint16(1), uint16(math.MaxUint16),
		uint32(0), uint32(1), uint32(math.MaxUint32),
		uint64(0), uint64(1), uint64(math.MaxUint64),
		int(1), int(math.MinInt64), int(math.MaxInt64),
		uint(0), uint(1), uint(math.MaxUint64),
		uintptr(0), uintptr(1), uintptr(math.MaxUint64),
	}
	for _, x := range values {
		want := ReflectInt(reflect.ValueOf(x))
		got := CastInt(x)
		assert.Equal(t, want, got)
	}
}
