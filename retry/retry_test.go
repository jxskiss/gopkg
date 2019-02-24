package retry

import (
	"fmt"
	"testing"
	"time"

	"github.com/jxskiss/errors"
	"github.com/stretchr/testify/assert"
)

func Test_Retry(t *testing.T) {
	is := assert.New(t)
	target := fakeErrors(3)

	begin := time.Now()
	r := Retry(3, 100*time.Millisecond, target)
	cost := time.Since(begin)
	is.True(!r.Ok)
	is.Equal(r.Attempts, 3)
	is.True(cost > 150*time.Millisecond)
	is.True(cost < 450*time.Millisecond)

	merr, ok := r.Error.(errors.ErrorGroup)
	is.True(ok)
	merrors := merr.Errors()
	is.Equal(len(merrors), 3)
	is.Equal(merrors[0].Error(), "error 3")
	is.Equal(merrors[1].Error(), "error 2")
	is.Equal(merrors[2].Error(), "error 1")

	target = fakeErrors(3)
	begin = time.Now()
	r = Retry(5, 100*time.Millisecond, target)
	cost = time.Since(begin)
	is.True(r.Ok)
	is.Equal(r.Attempts, 4)
	is.True(cost > 350*time.Millisecond)
	is.True(cost < 1050*time.Millisecond)

	merr, ok = r.Error.(errors.ErrorGroup)
	is.True(ok)
	merrors = merr.Errors()
	is.Equal(len(merrors), 3)
	is.Equal(merrors[0].Error(), "error 3")
	is.Equal(merrors[1].Error(), "error 2")
	is.Equal(merrors[2].Error(), "error 1")
}

func Test_Hook(t *testing.T) {
	is := assert.New(t)
	target := fakeErrors(3)
	hook := &fakeHook{}

	r := Retry(5, time.Millisecond, target, Hook(hook.log))
	is.True(r.Ok)
	is.Equal(hook.attempts, 3)
	is.Equal(len(hook.errs), 3)
	is.Equal(hook.errs[0], "error 1")
	is.Equal(hook.errs[1], "error 2")
	is.Equal(hook.errs[2], "error 3")
}

func fakeErrors(errCount int) func() error {
	attempt := 0
	return func() error {
		attempt++
		if attempt <= errCount {
			return fmt.Errorf("error %d", attempt)
		}
		return nil
	}
}

type fakeHook struct {
	attempts int
	errs     []string
}

func (h *fakeHook) log(attempts int, err error) {
	h.attempts = attempts
	h.errs = append(h.errs, err.Error())
}
