package monkey_test

import (
	"math/rand"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jxskiss/gopkg/v2/unsafe/monkey"
	"github.com/jxskiss/gopkg/v2/unsafe/monkey/testpkg"
)

//go:noinline
func no() bool {
	x := rand.Intn(10)
	return x < 0
}

//go:noinline
func yes() bool {
	x := rand.Intn(10)
	return x+1 > 0
}

func TestPatchFunc_simple(t *testing.T) {
	assert.True(t, !no())
	patch := monkey.PatchFunc(no, yes)
	assert.True(t, no())
	patch.Delete()
	assert.True(t, !no())
}

func TestPatchFunc_timeNow(t *testing.T) {
	before := time.Now()
	patch := monkey.PatchFunc(time.Now, func() time.Time {
		return time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	})
	during := time.Now()

	patch.Delete()
	after := time.Now()

	assert.Equal(t, time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC), during)
	assert.NotEqual(t, before, during)
	assert.NotEqual(t, during, after)
	assert.Equal(t, before.Truncate(time.Second), after.Truncate(time.Second))
}

func TestPatchMultipleTimes(t *testing.T) {
	assert.Equal(t, "testpkg.a", testpkg.A())

	monkey.AutoUnpatch(func() {
		fn1 := func() string { return "fn1" }
		monkey.PatchFunc(testpkg.A, fn1)
		assert.Equal(t, "fn1", testpkg.A())

		fn2 := func() string { return "fn2" }
		monkey.PatchFunc(testpkg.A, fn2)
		assert.Equal(t, "fn2", testpkg.A())

		fn3 := func() string { return "fn3" }
		monkey.PatchFunc(testpkg.A, fn3)
		assert.Equal(t, "fn3", testpkg.A())
	})
}

func TestAutoUnpatch(t *testing.T) {
	value := true
	monkey.AutoUnpatch(func() {
		monkey.PatchFunc(no, func() bool {
			return value
		})
		runtime.GC()
		assert.True(t, no())
	})
	assert.False(t, no())
}

type s struct {
	value bool
}

func (s *s) yes() bool { return s.value }

func TestPatchFunc_toInstanceMethod(t *testing.T) {
	i1 := &s{value: true}

	assert.True(t, !no())
	patch1 := monkey.PatchFunc(no, i1.yes)
	assert.True(t, no())

	i2 := &s{value: false}
	patch2 := monkey.PatchFunc(no, i2.yes)
	assert.False(t, no())

	patch2.Delete()
	patch1.Delete()
}

type f struct{}

//go:noinline
func (f *f) No() bool {
	x := rand.Intn(10)
	return x < 0
}

func TestPatchMethod(t *testing.T) {
	i := &f{}
	assert.True(t, !i.No())
	patch := monkey.PatchMethod(i, "No", func(_ *f) bool { return true })
	assert.True(t, i.No())

	patch.Delete()
	assert.True(t, !i.No())
}

func TestNotFunction(t *testing.T) {
	assert.Panics(t, func() {
		monkey.PatchFunc(no, 1)
	})
	assert.Panics(t, func() {
		monkey.PatchFunc(1, yes)
	})
}

func TestNotCompatible(t *testing.T) {
	assert.Panics(t, func() {
		monkey.PatchFunc(no, func() {})
	})
}

func TestPatchByName(t *testing.T) {
	testpkg_a := "github.com/jxskiss/gopkg/v2/unsafe/monkey/testpkg.a"
	assert.Equal(t, "testpkg.a", testpkg.A())

	patch := monkey.PatchByName(testpkg_a, func() string { return "TestPatchByName" })
	assert.Equal(t, "TestPatchByName", testpkg.A())

	patch.Delete()
	assert.Equal(t, "testpkg.a", testpkg.A())
}
