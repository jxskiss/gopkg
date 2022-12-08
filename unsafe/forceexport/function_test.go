package forceexport

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jxskiss/gopkg/v2/unsafe/forceexport/testpkg"
)

func TestGetPrivateFunc(t *testing.T) {
	var pi = 3.14
	got1 := testpkg.Floor(pi)
	assert.Equal(t, 3.0, got1)

	var privateFloorFunc func(x float64) float64
	assert.NotPanics(t, func() {
		GetFunc(&privateFloorFunc, "github.com/jxskiss/gopkg/v2/unsafe/forceexport/testpkg.floor")
	})
	got2 := privateFloorFunc(pi)

	assert.Equal(t, got1, got2)
}

// Note that we need to disable inlining here, or else the function won't be
// compiled into the binary. We also need to call it from the test so that the
// compiler doesn't remove it because it's unused.
//
//go:noinline
func addOne(x int) int {
	return x + 1
}

func TestAddOne(t *testing.T) {
	assert.Equal(t, 4, addOne(3))

	var addOneFunc func(x int) int
	assert.NotPanics(t, func() {
		GetFunc(&addOneFunc, "github.com/jxskiss/gopkg/v2/unsafe/forceexport.addOne")
	})
	assert.Equal(t, 4, addOneFunc(3))
}

func TestGetSelf(t *testing.T) {
	var getFunc func(interface{}, string)
	assert.NotPanics(t, func() {
		GetFunc(&getFunc, "github.com/jxskiss/gopkg/v2/unsafe/forceexport.GetFunc")
	})

	_p := func(fn interface{}) string { return fmt.Sprintf("%p", fn) }

	// The two functions should share the same code pointer, so they should
	// have the same string representation.
	assert.Equal(t, _p(getFunc), _p(GetFunc))

	// Call it again on itself!
	assert.NotPanics(t, func() {
		getFunc(&getFunc, "github.com/jxskiss/gopkg/v2/unsafe/forceexport.GetFunc")
	})
	assert.Equal(t, _p(getFunc), _p(GetFunc))
}

func TestInvalidFunc(t *testing.T) {
	var invalidFunc func()
	assert.Panics(t, func() {
		GetFunc(&invalidFunc, "invalidpackage.invalidfunction")
	})
	assert.Nil(t, invalidFunc)
}
