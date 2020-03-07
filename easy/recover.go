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
	return fmt.Sprintf("catch panic: %v, location: %v, check PanicError.Stack for stacktrace", p.Err, p.Loc)
}

func Go(f func(), logger ...interface{}) {
	err := Safe(f)()
	if err != nil && len(logger) > 0 {
		if perr, ok := err.(*PanicError); ok {
			logErr(logger[0], "%s\nstacktrace:\n%s", perr.Error(), perr.Stack)
		} else {
			logErr(logger[0], "catch error: %v", err)
		}
	}
}

func Go1(f func() error, logger ...interface{}) {
	err := Safe1(f)()
	if err != nil && len(logger) > 0 {
		if perr, ok := err.(*PanicError); ok {
			logErr(logger[0], "%s\nstacktrace:\n%s", perr.Error(), perr.Stack)
		} else {
			logErr(logger[0], "catch error: %v", err)
		}
	}
}

func Safe(f func()) func() error {
	return func() (err error) {
		defer func() {
			e := recover()
			if e == nil {
				return
			}
			panicLoc := identifyPanicLoc()
			err = &PanicError{
				Err:   EnsureError(e),
				Loc:   panicLoc,
				Stack: debug.Stack(),
			}
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
			panicLoc := identifyPanicLoc()
			err = &PanicError{
				Err:   EnsureError(e),
				Loc:   panicLoc,
				Stack: debug.Stack(),
			}
		}()
		err = f()
		return
	}
}

func identifyPanicLoc() string {
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
