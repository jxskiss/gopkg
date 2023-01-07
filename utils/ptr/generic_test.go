package ptr

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopy(t *testing.T) {
	var x = 1234
	y := Copy(&x)
	assert.NotEqual(t, fmt.Sprintf("%p", &x), fmt.Sprintf("%p", y))
	assert.Equal(t, x, *y)
}

func TestDeref(t *testing.T) {
	var x *int64
	assert.Equal(t, int64(0), Deref(x))

	x = Ptr(int64(1234))
	assert.Equal(t, int64(1234), Deref(x))
}

func TestPtr(t *testing.T) {
	var x = 1234
	y := Ptr(x)
	assert.Equal(t, x, Deref(y))
}

func TestNotZero(t *testing.T) {
	ret1 := NotZero(0)
	assert.Nil(t, ret1)

	ret2 := NotZero(1234)
	assert.NotNil(t, ret2)
}
