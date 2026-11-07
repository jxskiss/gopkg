package reflectx

import (
	"testing"
	"unsafe"

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
		{(chan int)(nil), true},
		{(func())(nil), true},
		{unsafe.Pointer(nil), true},
		{map[string]int{}, false},
		{[]string{}, false},
		{1, false},
		{"abc", false},
		{simple{}, false},
		{&simple{}, false},
		{struct{ P *int }{}, false},
		{[1]*int{}, false},
		{struct{ M map[string]int }{}, false},
		{[1]chan int{}, false},
		{make(chan int), false},
		{func() {}, false},
		{unsafe.Pointer(new(int)), false},
	}
	for i, tc := range testcases {
		got := IsNil(tc.v)
		assert.Equalf(t, tc.want, got, "i= %v, v = %q", i, tc.v)
	}
}
