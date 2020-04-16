package easy

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

type PanicError struct {
	Err   error
	Loc   string
	Stack []byte
}

func (p *PanicError) Error() string {
	return fmt.Sprintf("catch panic: %v, location: %v", p.Err, p.Loc)
}

func Go(f func(), logger ...interface{}) {
	go func() {
		err := Safe(f)()
		if err != nil && len(logger) > 0 {
			perr := err.(*PanicError)
			logError(logger[0], "%s\n%s", perr.Error(), perr.Stack)
		}
	}()
}

func Go1(f func() error, logger ...interface{}) {
	go func() {
		err := Safe1(f)()
		if err != nil && len(logger) > 0 {
			if perr, ok := err.(*PanicError); ok {
				logError(logger[0], "%s\n%s", perr.Error(), perr.Stack)
			} else {
				logError(logger[0], "catch error: %v", err)
			}
		}
	}()
}

func Safe(f func()) func() error {
	return func() (err error) {
		defer func() {
			e := recover()
			if e == nil {
				return
			}
			panicLoc := IdentifyPanic()
			err = EnsureError(e)
			err = &PanicError{Err: err, Loc: panicLoc, Stack: debug.Stack()}
		}()
		f()
		return nil
	}
}

func Safe1(f func() error) func() error {
	return func() (err error) {
		defer func() {
			e := recover()
			if e == nil {
				return
			}
			panicLoc := IdentifyPanic()
			err = EnsureError(e)
			err = &PanicError{Err: err, Loc: panicLoc, Stack: debug.Stack()}
		}()
		err = f()
		return
	}
}

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

func EnsureError(v interface{}) error {
	err, ok := v.(error)
	if !ok {
		err = fmt.Errorf("%v", v)
	}
	return err
}

func PanicOnError(args ...interface{}) {
	for _, arg := range args {
		if err, ok := arg.(error); ok && err != nil {
			panic(err)
		}
	}
}

var Must = PanicOnError
