package easy

import (
	"context"
	"reflect"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

type stringerFunc func(v interface{}) string

/*
DEBUG is debug message logger which do nothing if debug level is not enabled (the default).
It gives best performance for production deployment by eliminating unnecessary
parameter evaluation and control flows.

For sake of best performance, you may add "release" tag to the `go build` command,
then the debug calling is all empty functions, which will be eliminated by the
compiler and won't be compiled into the final binary file. It happens at compile-time.

DEBUG accepts very flexible arguments to help development, see the following examples:

	// func()
	DEBUG(func() {
		logrus.Debug(something)
		for _, item := range someSlice {
			logrus.Debug(item)
		}
	})

	// logger from context with or without format
	DEBUG(ctx, 1, 2, 3)
	DEBUG(ctx, "a=%v b=%v c=%v", 1, 2, 3)

	// log variables with or without format
	DEBUG(1, 2, 3)
	DEBUG("a=%v b=%v c=%v", 1, 2, 3)

	// optionally a DebugLogger can be provided as the first parameter
	logger := logrus.WithField("key", "dummy")
	DEBUG(logger, "aa", 2, 3)
	DEBUG(logger, "a=%v b=%v c=%v", "aa", 2, 3)

	// or a function which returns a logger implements DebugLogger
	logger := func() *logrus.Entry { ... }
	DEBUG(logger, "aa", 2, 3)
	DEBUG(logger, "a=%v b=%v c=%v", "aa", 2, 3)

	// or a print function to print the log message
	logger := logrus.Debugf
	DEBUG(logger, "aa", 2, 3)
	DEBUG(logger, "a=%v b=%v c=%v", "aa", 2, 3)

	// non-basic-type values will be formatted using json
	obj := &SomeStructType{Field1: "blah", Field2: 1234567, Field3: true}
	DEBUG(logger, "obj=%v"， obj)
*/
func DEBUG(args ...interface{}) {
	stringer := JSON
	logdebug(1, stringer, args...)
}

// DEBUGSkip is similar to DEBUG, but it has an extra skip param to skip stacktrace
// to get correct caller information.
// When you wrap functions in this package, you always want to use the functions
// which end with "Skip".
func DEBUGSkip(skip int, args ...interface{}) {
	stringer := JSON
	logdebug(skip+1, stringer, args...)
}

// PRETTY is similar to DEBUG, but it calls Pretty to format non-basic-type data.
func PRETTY(args ...interface{}) {
	stringer := Pretty
	logdebug(1, stringer, args...)
}

// PRETTYSkip is similar to PRETTY, but it has an extra skip param to skip stacktrace
// to get correct caller information.
// When you wrap functions in this package, you always want to use the functions
// which end with "Skip".
func PRETTYSkip(skip int, args ...interface{}) {
	stringer := Pretty
	logdebug(skip+1, stringer, args...)
}

// SPEW is similar to DEBUG, but it calls spew.Sprintf to format non-basic-type data.
func SPEW(args ...interface{}) {
	stringer := func(v interface{}) string { return spew.Sprintf("%#v", v) }
	logdebug(1, stringer, args...)
}

// SPEWSkip is similar to SPEW, but it has an extra skip param to skip stacktrace
// to get correct caller information.
// When you wrap functions in this package, you always want to use the functions
// which end with "Skip".
func SPEWSkip(skip int, args ...interface{}) {
	stringer := func(v interface{}) string { return spew.Sprintf("%#v", v) }
	logdebug(skip+1, stringer, args...)
}

// DUMP is similar to DEBUG, but it calls spew.Sdump to format non-basic-type data.
func DUMP(args ...interface{}) {
	stringer := func(v interface{}) string { return spew.Sdump(v) }
	logdebug(1, stringer, args...)
}

// DUMPSkip is similar to DUMP, but it has an extra skip param to skip stacktrace
// to get correct caller information.
// When you wrap functions in this package, you always want to use the functions
// which end with "Skip".
func DUMPSkip(skip int, args ...interface{}) {
	stringer := func(v interface{}) string { return spew.Sdump(v) }
	logdebug(skip+1, stringer, args...)
}

func logdebug(skip int, stringer stringerFunc, args ...interface{}) {
	if _logcfg.EnableDebug == nil || !_logcfg.EnableDebug() {
		return
	}
	var logger DebugLogger
	if len(args) > 0 {
		if arg0, ok := args[0].(func()); ok {
			arg0()
			return
		}
		logger, args = parseLogger(args)
	}
	if logger == nil {
		logger = _logcfg.getLogger(nil)
	}
	outputDebugLog(skip+1, logger, stringer, args)
}

func outputDebugLog(skip int, logger DebugLogger, stringer stringerFunc, args []interface{}) {
	caller, file, line := Caller(skip + 1)
	callerPrefix := "[" + caller + "] "
	if len(args) > 0 {
		if format, ok := args[0].(string); ok && strings.IndexByte(format, '%') >= 0 {
			logger.Debugf(callerPrefix+format, formatArgs(stringer, args[1:])...)
			return
		}
		format := callerPrefix + "%v" + strings.Repeat(" %v", len(args)-1)
		logger.Debugf(format, formatArgs(stringer, args)...)
	} else {
		logger.Debugf("========  %s#L%d - %s  ========", file, line, caller)
	}
}

var debugLoggerTyp = reflect.TypeOf((*DebugLogger)(nil)).Elem()

func parseLogger(args []interface{}) (DebugLogger, []interface{}) {
	var logger DebugLogger
	if arg0, ok := args[0].(DebugLogger); ok {
		logger = arg0
		args = args[1:]
		return logger, args
	}

	switch arg0 := args[0].(type) {
	case context.Context:
		logger = _logcfg.getLogger(&arg0)
		args = args[1:]
	case func(string, ...interface{}):
		logger = PrintFunc(arg0)
		args = args[1:]
	default:
		arg0typ := reflect.TypeOf(arg0)
		if arg0typ.Kind() == reflect.Func {
			if arg0typ.NumIn() == 0 && arg0typ.NumOut() == 1 &&
				arg0typ.Out(0).Implements(debugLoggerTyp) {
				out := reflect.ValueOf(arg0).Call(nil)[0]
				logger = out.Interface().(DebugLogger)
			}
			args = args[1:]
		}
	}
	return logger, args
}
