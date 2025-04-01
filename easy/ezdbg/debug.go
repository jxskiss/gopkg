package ezdbg

import (
	"context"
	"reflect"
	"runtime"
	"strings"
	"unicode/utf8"

	"github.com/jxskiss/gopkg/v2/easy"
)

type stringerFunc func(v any) string

/*
DEBUG is debug message logger which do nothing if debug level is not enabled (the default).
It has good performance for production deployment by eliminating unnecessary
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

func logdebug(skip int, stringer stringerFunc, args ...any) {
	ctx := context.Background()
	if len(args) > 0 {
		if arg0ctx, ok := args[0].(context.Context); ok {
			if arg0ctx != nil {
				ctx = arg0ctx
			}
			args = args[1:]
		}
	}
	if _logcfg.EnableDebug == nil || !_logcfg.EnableDebug(ctx) {
		return
	}

	// Check filter rules.
	caller, fullFileName, simpleFileName, line := getCaller(skip + 1)
	if _logcfg.filter != nil && !_logcfg.filter.Allow(fullFileName) {
		return
	}

	var logger DebugLogger
	if len(args) > 0 {
		if arg0, ok := args[0].(func()); ok {
			arg0()
			return
		}
		logger, args = parseArg0Logger(args)
	}
	if logger == nil {
		logger = _logcfg.getLogger(ctx)
	}
	callerPrefix := "[" + caller + "] "
	if len(args) > 0 {
		if format, ok := args[0].(string); ok && strings.IndexByte(format, '%') >= 0 {
			logger.Debugf(callerPrefix+format, formatArgs(stringer, args[1:])...)
			return
		}
		format := callerPrefix + "%v" + strings.Repeat(" %v", len(args)-1)
		logger.Debugf(format, formatArgs(stringer, args)...)
	} else {
		logger.Debugf("========  %s#L%d - %s  ========", simpleFileName, line, caller)
	}
}

var debugLoggerTyp = reflect.TypeOf((*DebugLogger)(nil)).Elem()

func parseArg0Logger(args []any) (DebugLogger, []any) {
	var logger DebugLogger
	switch arg0 := args[0].(type) {
	case DebugLogger:
		logger = arg0
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

func getCaller(skip int) (funcName, fullFileName, simpleFileName string, line int) {
	pc, fullFileName, line, _ := runtime.Caller(skip + 1)
	funcName = runtime.FuncForPC(pc).Name()
	for i := len(funcName) - 1; i >= 0; i-- {
		if funcName[i] == '/' {
			funcName = funcName[i+1:]
			break
		}
	}
	simpleFileName = fullFileName
	pathSepCnt := 0
	for i := len(simpleFileName) - 1; i >= 0; i-- {
		if simpleFileName[i] == '/' {
			pathSepCnt++
			if pathSepCnt == 2 {
				simpleFileName = simpleFileName[i+1:]
				break
			}
		}
	}
	return
}
