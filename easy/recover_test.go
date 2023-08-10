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

func TestRecover(t *testing.T) {
	var got string
	recoverfn := Recover(func(ctx context.Context, panicErr *PanicError) {
		got = fmt.Sprintf("%+v", panicErr)
	})
	func() {
		defer recoverfn(context.Background())
		willPanic()
	}()
	t.Log(got)
	assert.Contains(t, got, "panic: oops..., location: github.com/jxskiss/"+wantPanicLoc)
	assert.Contains(t, got, "gopkg/easy/recover.go:94")
	assert.Contains(t, got, "gopkg/easy/recover_test.go:20")
	assert.Contains(t, got, "gopkg/easy/recover_test.go:34")
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
