package easy

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

var _ = log.Println

var wantPanicLoc string

func willPanic() {
	wantPanicLoc = "gopkg/v2/easy.willPanic:20"
	panic("oops...")
}

func willPanicCaller() {
	willPanic()
}

func TestSafe(t *testing.T) {
	var err error
	assert.NotPanics(t, func() {
		err = Safe(func() {
			willPanic()
		})()
	})
	assert.NotNil(t, err)
	assert.IsType(t, &PanicError{}, err)

	got := fmt.Sprintf("%+v", err)
	t.Log(got)
	assert.Contains(t, got, "panic: oops..., location: github.com/jxskiss/"+wantPanicLoc)
}

func TestSafe1(t *testing.T) {
	var err error
	assert.NotPanics(t, func() {
		err = Safe1(func() error {
			willPanic()
			return errors.New("test error")
		})()
	})
	assert.NotNil(t, err)
	assert.IsType(t, &PanicError{}, err)

	got := fmt.Sprintf("%+v", err)
	t.Log(got)
	assert.Contains(t, got, "panic: oops..., location: github.com/jxskiss/"+wantPanicLoc)
}

func TestRecover(t *testing.T) {
	var err error
	assert.NotPanics(t, func() {
		func() {
			defer Recover(&err)
			willPanic()
		}()
	})
	assert.NotNil(t, err)

	got := fmt.Sprintf("%+v", err)
	assert.Contains(t, got, "panic: oops..., location: github.com/jxskiss/"+wantPanicLoc)
}

func TestNewRecoverFunc(t *testing.T) {
	var got string
	recoverfn := NewRecoverFunc(func(ctx context.Context, panicErr *PanicError) {
		got = fmt.Sprintf("%+v", panicErr)
	})
	func() {
		defer recoverfn(context.Background(), nil)
		willPanic()
	}()
	t.Log(got)
	assert.Contains(t, got, "panic: oops..., location: github.com/jxskiss/"+wantPanicLoc)
	assert.Contains(t, got, "gopkg/easy/recover_test.go:20  (github.com/jxskiss/gopkg/v2/easy.willPanic)")
	assert.Contains(t, got, "gopkg/easy/recover_test.go:79  (github.com/jxskiss/gopkg/v2/easy.TestNewRecoverFunc.func2)")
	assert.Contains(t, got, "gopkg/easy/recover_test.go:80  (github.com/jxskiss/gopkg/v2/easy.TestNewRecoverFunc)")
}

func TestIdentifyPanicLoc(t *testing.T) {
	var panicLoc1 string
	func() {
		defer func() {
			recover()
			panicLoc1 = IdentifyPanic()
		}()
		willPanic()
	}()
	t.Log(panicLoc1)
	assert.True(t, strings.HasSuffix(panicLoc1, wantPanicLoc))

	var panicLoc2 string
	func() {
		defer func() {
			recover()
			panicLoc2 = IdentifyPanic()
		}()
		willPanicCaller()
	}()
	t.Log(panicLoc2)
	assert.True(t, strings.HasSuffix(panicLoc2, wantPanicLoc))
}

func TestPanicOnError(t *testing.T) {
	panicErr := errors.New("dummy panic error")
	willPanic := func() (int, error) {
		return 123, panicErr
	}

	x, gotErr := willPanic()
	assert.PanicsWithValue(t, panicErr, func() {
		PanicOnError(x, gotErr)
	})

	assert.PanicsWithValue(t, panicErr, func() {
		PanicOnError(willPanic())
	})
}
