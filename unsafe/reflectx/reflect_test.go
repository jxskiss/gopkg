package reflectx

import (
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
