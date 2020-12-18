package monkey_test

import (
	"runtime"
	"testing"
	"time"

	"github.com/jxskiss/gopkg/monkey"
	"github.com/jxskiss/gopkg/monkey/testpkg"
	"github.com/stretchr/testify/assert"
)

//go:noinline
func no() bool { return false }

//go:noinline
func yes() bool { return true }

func TestTimePatch(t *testing.T) {
	before := time.Now()
	monkey.Patch(time.Now, func() time.Time {
		return time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)
	})
	during := time.Now()
	assert.True(t, monkey.Unpatch(time.Now))

	after := time.Now()
	assert.Equal(t, time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC), during)
	assert.NotEqual(t, before, during)
	assert.NotEqual(t, during, after)
}

func TestGC(t *testing.T) {
	value := true
	monkey.Patch(no, func() bool {
		return value
	})
	defer monkey.UnpatchAll()
	runtime.GC()
	assert.True(t, no())
}

func TestSimple(t *testing.T) {
	assert.True(t, !no())
	monkey.Patch(no, yes)
	assert.True(t, no())
	assert.True(t, monkey.Unpatch(no))
	assert.True(t, !no())
	assert.True(t, !monkey.Unpatch(no))
}

func TestPatchGuard(t *testing.T) {
	var patch *monkey.PatchGuard
	patch = monkey.Patch(no, yes)
	assert.True(t, no())

	patch = monkey.Patch(no, func() bool {
		patch.Unpatch()
		defer patch.Restore()
		return !no()
	})
	for i := 0; i < 100; i++ {
		assert.True(t, no())
	}
	assert.True(t, no())
	assert.True(t, monkey.Unpatch(no))
}

func TestUnpatchAll(t *testing.T) {
	assert.True(t, !no())
	monkey.Patch(no, yes)
	assert.True(t, no())
	monkey.UnpatchAll()
	assert.True(t, !no())
}

type s struct{}

func (s *s) yes() bool { return true }

func TestWithInstanceMethod(t *testing.T) {
	i := &s{}

	assert.True(t, !no())
	monkey.Patch(no, i.yes)
	assert.True(t, no())
	monkey.Unpatch(no)
	assert.True(t, !no())
}

type f struct{}

//go:noinline
func (f *f) No() bool { return false }

func TestOnInstanceMethod(t *testing.T) {
	i := &f{}
	assert.True(t, !i.No())
	monkey.PatchMethod(i, "No", func(_ *f) bool { return true })
	assert.True(t, i.No())
	assert.True(t, monkey.UnpatchMethod(i, "No"))
	assert.True(t, !i.No())
}

func TestNotFunction(t *testing.T) {
	assert.Panics(t, func() {
		monkey.Patch(no, 1)
	})
	assert.Panics(t, func() {
		monkey.Patch(1, yes)
	})
}

func TestNotCompatible(t *testing.T) {
	assert.Panics(t, func() {
		monkey.Patch(no, func() {})
	})
}

func TestPatchByTargetName(t *testing.T) {
	testpkg_a := "github.com/jxskiss/gopkg/monkey/testpkg.a"
	assert.Equal(t, "testpkg.a", testpkg.A())
	monkey.PatchByName(testpkg_a, func() string { return "TestPatchByTargetName" })
	assert.Equal(t, "TestPatchByTargetName", testpkg.A())

	assert.True(t, monkey.UnpatchByName(testpkg_a))
	assert.Equal(t, "testpkg.a", testpkg.A())
}
