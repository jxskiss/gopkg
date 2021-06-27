package easy

import (
	"context"
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
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

// Go calls the given function with panic recover, in case of panic happens,
// the panic message, location and the calling stack will be logged using
// the default logger configured by `ConfigLog` in this package.
func Go(f func()) {
	go func() {
		defer Recover(nil, nil)
		f()
	}()
}

// Go1 calls the given function with panic recover, in case an error is returned,
// or panic happens, the error message or panic information will be logged
// using the default logger configured by `ConfigLog` in this package.
func Go1(f func() error) {
	go func() {
		defer Recover(nil, nil)
		err := f()
		if err == nil {
			return
		}
		_logcfg.getLogger(nil).Errorf("catch error: %v", err)
	}()
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
			panicLoc := IdentifyPanic()
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
			panicLoc := IdentifyPanic()
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

// Recover recovers unexpected panic, and log error messages using
// logger associated with the given context, if `err` is not nil,
// an wrapped `PanicError` will be assigned to it.
//
// Note that this function should not be wrapped be another function,
// instead it should be called directly by the `defer` statement,
// or it won't work as you may expect.
func Recover(ctx context.Context, err *error) {
	e := recover()
	if e == nil {
		return
	}

	panicLoc := IdentifyPanic()
	stack := debug.Stack()
	pErr := &PanicError{
		Exception:  e,
		Location:   panicLoc,
		Stacktrace: stack,
	}

	// If the caller receives the error, we don't log it here,
	// else we log the panic error with stack information.
	if err != nil {
		*err = pErr
		return
	}
	_logcfg.getLogger(&ctx).Errorf("catch %v\n%s", pErr, stack)
}

// IdentifyPanic reports the panic location when a panic happens.
func IdentifyPanic() string {
	var name, file string
	var line int
	var pc [16]uintptr

	n := runtime.Callers(3, pc[:])
	for _, pc := range pc[:n] {
		fn := runtime.FuncForPC(pc)
		if fn == nil {
			continue
		}
		file, line = fn.FileLine(pc)
		name = fn.Name()
		if !strings.HasPrefix(name, "runtime.") {
			break
		}
	}
	switch {
	case name != "":
		return fmt.Sprintf("%v:%v", name, line)
	case file != "":
		return fmt.Sprintf("%v:%v", file, line)
	}

	return fmt.Sprintf("pc:%x", pc)
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
