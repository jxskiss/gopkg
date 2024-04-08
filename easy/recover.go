package easy

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"runtime"

	"github.com/jxskiss/gopkg/v2/internal"
)

var recoverToError = NewRecoverFunc(func(_ context.Context, _ *PanicError) {})

// PanicError represents an captured panic error.
type PanicError struct {
	Exception  any
	Location   string
	Stacktrace []byte
}

func newPanicError(skip int, e any) *PanicError {
	panicLoc, frames := internal.IdentifyPanic(skip + 1)
	stack := formatFrames(frames)
	return &PanicError{
		Exception:  e,
		Location:   panicLoc,
		Stacktrace: stack,
	}
}

func (p *PanicError) Unwrap() error {
	if err, ok := p.Exception.(error); ok {
		return err
	}
	return fmt.Errorf("%v", p.Exception)
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

// Safe returns a wrapper function with panic recover.
//
// Note that if panic happens, the wrapped function does not log messages,
// instead it will be returned as a `*PanicError`, the caller take
// responsibility to log the panic messages.
func Safe(f func()) func() error {
	return func() (err error) {
		defer recoverToError(context.Background(), &err)
		f()
		return
	}
}

// Safe1 returns a wrapper function with panic recover.
//
// Note that if panic or error happens, the wrapped function does not log
// messages, instead it will be returned as an error, the caller take
// responsibility to log the panic or error messages.
func Safe1(f func() error) func() error {
	return func() (err error) {
		defer recoverToError(context.Background(), &err)
		err = f()
		return
	}
}

// Recover recovers panics.
// If panic occurred, it prints an error log.
// If err is not nil, it will be set to a `*PanicError`.
//
// Note that this function should not be wrapped by another function,
// instead it should be called directly by the `defer` statement,
// else it won't work as you may expect.
func Recover(errp *error) {
	e := recover()
	if e == nil {
		return
	}
	pErr := newPanicError(0, e)
	if errp != nil {
		*errp = pErr
	}
	log.Printf("[Error] %+v", pErr)
}

// NewRecoverFunc returns a function which recovers panics.
// It accepts a panicErr handler function which may be used to log the
// panic error and context information.
//
// Note that the returned function should not be wrapped by another
// function, instead it should be called directly by the `defer` statement,
// else it won't work as you may expect.
//
// Example:
//
//	var Recover = NewRecoverFunc(func(ctx context.Context, panicErr *easy.PanicError) {
//		serviceName := getServiceName(ctx)
//
//		// emit metrics
//		metrics.Emit("panic", 1, metrics.Tag("service", serviceName))
//
//		// print log
//		log.Printf("[Error] %+v", panicErr)
//
//		// or check the panic details
//		// mylog.Logger(ctx).Errorf("catch panic: %v\nlocation: %s\nstacktrace: %s",
//		//	panicErr.Exception, panicErr.Location, panicErr.Stacktrace)
//	})
//
//	// Use the recover function somewhere.
//	func SomeFunction(ctx context.Context) (err error) {
//		defer Recover(ctx, &err)
//		// do something ...
//	}
func NewRecoverFunc[T any](f func(ctx T, panicErr *PanicError)) func(ctx T, errp *error) {
	return func(ctx T, errp *error) {
		e := recover()
		if e == nil {
			return
		}
		pErr := newPanicError(0, e)
		if errp != nil {
			*errp = pErr
		}
		f(ctx, pErr)
	}
}

// IdentifyPanic reports the panic location when a panic happens.
// It should be called directly after `recover()`, not wrapped by
// another function, else it returns incorrect location.
// Use IdentifyPanicSkip for wrapping.
func IdentifyPanic() string {
	loc, _ := internal.IdentifyPanic(1)
	return loc
}

// IdentifyPanicSkip is similar to IdentifyPanic, except that
// it accepts a param skip for wrapping usecase.
func IdentifyPanicSkip(skip int) string {
	loc, _ := internal.IdentifyPanic(skip + 1)
	return loc
}

// EnsureError ensures the given value (should be non-nil) is an error.
// If it's not an error, `fmt.Errorf("%v", v)` will be used to convert it.
func EnsureError(v any) error {
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
func PanicOnError(args ...any) {
	for _, arg := range args {
		if err, ok := arg.(error); ok && err != nil {
			panic(err)
		}
	}
}

func formatFrames(frames []runtime.Frame) []byte {
	var buf bytes.Buffer
	for _, f := range frames {
		file, line, funcName := f.File, f.Line, f.Function
		if file == "" {
			file = "unknown"
		}
		if funcName == "" {
			funcName = "unknown"
		}
		if buf.Len() > 0 {
			buf.WriteByte('\n')
		}
		fmt.Fprintf(&buf, "%s:%d  (%s)", file, line, funcName)
	}
	return buf.Bytes()
}
