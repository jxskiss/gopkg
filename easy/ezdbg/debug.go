package ezdbg

import (
	"context"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/davecgh/go-spew/spew"

	"github.com/jxskiss/gopkg/v2/easy"
)

type stringerFunc func(v any) string

/*
DEBUG is debug message logger which do nothing if debug level is not enabled (the default).
It gives best performance for production deployment by eliminating unnecessary
parameter evaluation and control flows.

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
func DEBUG(args ...any) {
	stringer := easy.JSON
	logdebug(1, stringer, args...)
}

// DEBUGSkip is similar to DEBUG, but it has an extra skip param to skip stacktrace
// to get correct caller information.
// When you wrap functions in this package, you always want to use the functions
// which end with "Skip".
func DEBUGSkip(skip int, args ...any) {
	stringer := easy.JSON
	logdebug(skip+1, stringer, args...)
}

// PRETTY is similar to DEBUG, but it calls Pretty to format non-basic-type data.
func PRETTY(args ...any) {
	stringer := easy.Pretty
	logdebug(1, stringer, args...)
}

// PRETTYSkip is similar to PRETTY, but it has an extra skip param to skip stacktrace
// to get correct caller information.
// When you wrap functions in this package, you always want to use the functions
// which end with "Skip".
func PRETTYSkip(skip int, args ...any) {
	stringer := easy.Pretty
	logdebug(skip+1, stringer, args...)
}

// SPEW is similar to DEBUG, but it calls spew.Sprintf to format non-basic-type data.
func SPEW(args ...any) {
	stringer := func(v any) string { return spew.Sprintf("%#v", v) }
	logdebug(1, stringer, args...)
}

// SPEWSkip is similar to SPEW, but it has an extra skip param to skip stacktrace
// to get correct caller information.
// When you wrap functions in this package, you always want to use the functions
// which end with "Skip".
func SPEWSkip(skip int, args ...any) {
	stringer := func(v any) string { return spew.Sprintf("%#v", v) }
	logdebug(skip+1, stringer, args...)
}

// DUMP is similar to DEBUG, but it calls spew.Sdump to format non-basic-type data.
func DUMP(args ...any) {
	stringer := func(v any) string { return spew.Sdump(v) }
	logdebug(1, stringer, args...)
}

// DUMPSkip is similar to DUMP, but it has an extra skip param to skip stacktrace
// to get correct caller information.
// When you wrap functions in this package, you always want to use the functions
// which end with "Skip".
func DUMPSkip(skip int, args ...any) {
	stringer := func(v any) string { return spew.Sdump(v) }
	logdebug(skip+1, stringer, args...)
}

func logdebug(skip int, stringer stringerFunc, args ...any) {
	ctx := context.Background()
	if len(args) > 0 {
		if _ctx, ok := args[0].(context.Context); ok && _ctx != nil {
			ctx = _ctx
		}
	}
	if _logcfg.EnableDebug == nil || !_logcfg.EnableDebug(ctx) {
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

func outputDebugLog(skip int, logger DebugLogger, stringer stringerFunc, args []any) {
	caller, file, line := easy.Caller(skip + 1)
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

func parseLogger(args []any) (DebugLogger, []any) {
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
	case func(string, ...any):
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

func formatArgs(stringer stringerFunc, args []any) []any {
	retArgs := make([]any, 0, len(args))
	for _, v := range args {
		x := v
		if v != nil {
			typ := reflect.TypeOf(v)
			if typ.Kind() == reflect.Ptr && !reflect.ValueOf(v).IsNil() && isBasicType(typ.Elem()) {
				typ = typ.Elem()
				v = reflect.ValueOf(v).Elem().Interface()
			}
			if isBasicType(typ) {
				x = v
			} else if bv, ok := v.([]byte); ok && utf8.Valid(bv) {
				x = string(bv)
			} else {
				x = stringer(v)
			}
		}
		retArgs = append(retArgs, x)
	}
	return retArgs
}

func isBasicType(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Bool, reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128:
		return true
	}
	return false
}
