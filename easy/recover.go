package easy

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/jxskiss/gopkg/v2/internal"
)

// PanicError represents an captured panic error.
type PanicError struct {
	Exception  interface{}
	Location   string
	Stacktrace []byte
}

func (p *PanicError) Error() string {
	return fmt.Sprintf("panic: %v, location: %v", p.Exception, p.Location)
}

func (p *PanicError) Format(f fmt.State, c rune) {
	if c == 'v' && f.Flag('+') {
		fmt.Fprintf(f, "panic: %v, location: %v\n%s\n", p.Exception, p.Location, p.Stacktrace)
	} else {
		fmt.Fprint(f, p.Error())
	}
}

// Safe returns an wrapped function with panic recover.
//
// Note that if panic happens, the wrapped function does not log messages,
// instead it will be returned as a `*PanicError`, the caller take
// responsibility to log the panic messages.
func Safe(f func()) func() error {
	return func() (err error) {
		defer func() {
			e := recover()
			if e == nil {
				return
			}
			panicLoc := internal.IdentifyPanic(0)
			stack := debug.Stack()
			err = &PanicError{
				Exception:  e,
				Location:   panicLoc,
				Stacktrace: stack,
			}
		}()
		f()
		return nil
	}
}

// Safe1 returns an wrapped function with panic recover.
//
// Note that if panic or error happens, the wrapped function does not log
// messages, instead it will be returned as an error, the caller take
// responsibility to log the panic or error messages.
func Safe1(f func() error) func() error {
	return func() (err error) {
		defer func() {
			e := recover()
			if e == nil {
				return
			}
			panicLoc := internal.IdentifyPanic(0)
			stack := debug.Stack()
			err = &PanicError{
				Exception:  e,
				Location:   panicLoc,
				Stacktrace: stack,
			}
		}()
		err = f()
		return
	}
}

// Recover returns a function which can be used to recover panics.
// It accepts a panicErr handler function which may be used to log the
// panic error and context information.
//
// Note that the returned function should not be wrapped by another
// function, instead it should be called directly by the `defer` statement,
// else it won't work as you may expect.
func Recover(f func(ctx context.Context, panicErr *PanicError)) func(ctx context.Context) {
	return func(ctx context.Context) {
		e := recover()
		if e == nil {
			return
		}
		panicLoc := internal.IdentifyPanic(0)
		stack := debug.Stack()
		pErr := &PanicError{
			Exception:  e,
			Location:   panicLoc,
			Stacktrace: stack,
		}
		f(ctx, pErr)
	}
}

// IdentifyPanic reports the panic location when a panic happens.
// It should be called directly after `recover()`, not wrapped by
// another function, else it returns incorrect location.
// Use IdentifyPanicSkip for wrapping.
func IdentifyPanic() string {
	return internal.IdentifyPanic(1)
}

// IdentifyPanicSkip is similar to IdentifyPanic, except that
// it accepts a param skip for wrapping usecase.
func IdentifyPanicSkip(skip int) string {
	return internal.IdentifyPanic(skip + 1)
}

// EnsureError ensures the given value (should be non-nil) is an error.
// If it's not an error, `fmt.Errorf("%v", v)` will be used to convert it.
func EnsureError(v interface{}) error {
	if v == nil {
		return nil
	}
	err, ok := v.(error)
	if !ok {
		err = fmt.Errorf("%v", v)
	}
	return err
}

// PanicOnError fires a panic if any of the args is non-nil error.
func PanicOnError(args ...interface{}) {
	for _, arg := range args {
		if err, ok := arg.(error); ok && err != nil {
			panic(err)
		}
	}
}
