package reflectx

// This file is modified from https://github.com/alangpierce/go-forceexport.

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTimeNow(t *testing.T) {
	var timeNowFunc func() (int64, int32)
	err := GetFunc(&timeNowFunc, "time.now")
	assert.Nil(t, err)
	sec, nsec := timeNowFunc()
	assert.Greater(t, sec, int64(0))
	assert.Greater(t, nsec, int32(0))
}

// Note that we need to disable inlining here, or else the function won't be
// compiled into the binary. We also need to call it from the test so that the
// compiler doesn't remove it because it's unused.
//go:noinline
func addOne(x int) int {
	return x + 1
}

func TestAddOne(t *testing.T) {
	assert.Equal(t, 4, addOne(3))

	var addOneFunc func(x int) int
	err := GetFunc(&addOneFunc, "github.com/jxskiss/gopkg/reflectx.addOne")
	assert.Nil(t, err)
	assert.Equal(t, 4, addOneFunc(3))
}

func TestGetSelf(t *testing.T) {
	var getFunc func(interface{}, string) error
	err := GetFunc(&getFunc, "github.com/jxskiss/gopkg/reflectx.GetFunc")
	assert.Nil(t, err)

	_p := func(fn interface{}) string { return fmt.Sprintf("%p", fn) }

	// The two functions should share the same code pointer, so they should
	// have the same string representation.
	assert.Equal(t, _p(getFunc), _p(GetFunc))

	// Call it again on itself!
	err = getFunc(&getFunc, "github.com/jxskiss/gopkg/reflectx.GetFunc")
	assert.Nil(t, err)
	assert.Equal(t, _p(getFunc), _p(GetFunc))
}

func TestInvalidFunc(t *testing.T) {
	var invalidFunc func()
	err := GetFunc(&invalidFunc, "invalidpackage.invalidfunction")
	assert.Error(t, err)
	assert.Nil(t, invalidFunc)
}
