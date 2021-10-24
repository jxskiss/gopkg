package zlog

import (
	"context"
	"fmt"
	"runtime"
	"sync"

	"go.uber.org/zap"
)

var (
	gL, gL_1 *zap.Logger
	gS, gS_1 *zap.SugaredLogger
	gP       *Properties

	setupOnce sync.Once
)

func init() {
	replaceGlobals(mustNewGlobalLogger(&Config{}))
}

// Properties records some information about a zap logger.
type Properties struct {
	functionKey string
	ctxFunc     CtxFunc
	level       atomicLevel
}

// GetLevel gets the logging level of the logger.
func (p *Properties) GetLevel() Level { return p.level.Level() }

// SetLevel modifies the logging level of the logger.
func (p *Properties) SetLevel(lvl Level) { p.level.SetLevel(lvl) }

// SetupGlobals setups the global loggers in this package and zap library.
// By default, global loggers are set with default configuration with info
// level and json format, you may use this function to change the default
// loggers.
//
// If redirectStdLog is true, it calls RedirectStdLog to redirect output
// from the standard library's package-global logger to the global logger
// configured by this function.
//
// This function must not be called more than once, else it panics.
// It should be called only in main function at program startup, library
// code shall not touch this.
func SetupGlobals(cfg *Config, redirectStdLog bool) {
	first := false
	setupOnce.Do(func() {
		first = true
		replaceGlobals(mustNewGlobalLogger(cfg))
		if redirectStdLog {
			RedirectStdLog()
		}
	})
	if !first {
		panic("SetupGlobals called more than once")
	}
}

func mustNewGlobalLogger(cfg *Config, opts ...zap.Option) (*zap.Logger, *Properties) {
	logger, props, err := New(cfg, opts...)
	if err != nil {
		panic(fmt.Sprintf("invalid config to initialize logger: %v", err))
	}
	return logger, props
}

func replaceGlobals(logger *zap.Logger, props *Properties) func() {
	oldL, oldP := gL, gP

	gL = logger
	gS = logger.Sugar()
	gP = props

	gL_1 = logger.WithOptions(zap.AddCallerSkip(1))
	gS_1 = logger.WithOptions(zap.AddCallerSkip(1)).Sugar()

	zap.ReplaceGlobals(logger)

	return func() {
		replaceGlobals(oldL, oldP)
	}
}

// RedirectStdLog redirects output from the standard library's package-global
// logger to the global logger in this package. It returns a function to
// restore the original behavior of the standard library.
func RedirectStdLog() func() { return zap.RedirectStdLog(L()) }

// GetLevel gets the global logging level.
func GetLevel() Level { return gP.GetLevel() }

// SetLevel modifies the global logging level on the fly.
// It's safe for concurrent use.
func SetLevel(lvl Level) { gP.SetLevel(lvl) }

// L returns the global Logger, which can be reconfigured with
// SetupGlobals.
func L() *zap.Logger { return gL }

// S returns the global SugaredLogger, which can be reconfigured with
// SetupGlobals.
func S() *zap.SugaredLogger { return gS }

// Sync flushes any buffered log entries.
func Sync() error {
	if err := L().Sync(); err != nil {
		return err
	}
	if err := S().Sync(); err != nil {
		return err
	}
	if err := _l().Sync(); err != nil {
		return err
	}
	if err := _s().Sync(); err != nil {
		return err
	}
	return nil
}

// -------- global logging functions -------- //

func _l() *zap.Logger        { return gL_1 }
func _s() *zap.SugaredLogger { return gS_1 }

func Trace(msg string, fields ...zap.Field) {
	if GetLevel() <= TraceLevel {
		_l().Debug(tracePrefix+msg, fields...)
	}
}

func Debug(msg string, fields ...zap.Field)  { _l().Debug(msg, fields...) }
func Info(msg string, fields ...zap.Field)   { _l().Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)   { _l().Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field)  { _l().Error(msg, fields...) }
func DPanic(msg string, fields ...zap.Field) { _l().DPanic(msg, fields...) }
func Panic(msg string, fields ...zap.Field)  { _l().Panic(msg, fields...) }
func Fatal(msg string, fields ...zap.Field)  { _l().Fatal(msg, fields...) }

func Tracef(format string, args ...interface{}) {
	if GetLevel() <= TraceLevel {
		_s().Debugf(tracePrefix+format, args...)
	}
}

func Debugf(format string, args ...interface{})  { _s().Debugf(format, args...) }
func Infof(format string, args ...interface{})   { _s().Infof(format, args...) }
func Warnf(format string, args ...interface{})   { _s().Warnf(format, args...) }
func Errorf(format string, args ...interface{})  { _s().Errorf(format, args...) }
func DPanicf(format string, args ...interface{}) { _s().DPanicf(format, args...) }
func Panicf(format string, args ...interface{})  { _s().Panicf(format, args...) }
func Fatalf(format string, args ...interface{})  { _s().Fatalf(format, args...) }

// -------- utility functions -------- //

// With creates a child logger and adds structured context to it.
// Fields added to the child don't affect the parent, and vice versa.
func With(fields ...zap.Field) *zap.Logger {
	return L().With(fields...)
}

// WithCtx creates a child logger and customizes its behavior using context
// data (e.g. adding fields, dynamically change logging level, etc.)
//
// If the ctx is created by WithBuilder, it carries a Builder instance,
// this function uses that Builder to build the logger, else it calls
// Config.CtxFunc to get CtxResult from ctx. In case Config.CtxFunc is
// not configured globally, it logs an error message at DPANIC level.
//
// Also see Config.CtxFunc, CtxArgs and CtxResult for more details.
func WithCtx(ctx context.Context, extra ...zap.Field) *zap.Logger {
	if ctx == nil {
		return L().With(extra...)
	}
	if builder := getCtxBuilder(ctx); builder != nil {
		return builder.With(extra...).Build()
	}
	ctxFunc := gP.ctxFunc
	if ctxFunc == nil {
		L().DPanic("calling WithCtx without CtxFunc configured")
		return L().With(extra...)
	}
	ctxResult := ctxFunc(ctx, CtxArgs{})
	return B(nil).withCtxResult(ctxResult).With(extra...).Build()
}

// WithMethod creates a child logger and adds the caller's method name
// to the logger if Config.FunctionKey is not configured.
// It will also add extra to the logger's fields.
func WithMethod(extra ...zap.Field) *zap.Logger {
	if gP.functionKey != "" {
		return L().With(extra...)
	}
	methodName, _, _, ok := getCaller(1)
	if !ok {
		return L().With(extra...)
	}
	if len(extra) == 0 {
		return L().With(zap.String(MethodKey, methodName))
	}
	fields := append([]zap.Field{zap.String(MethodKey, methodName)}, extra...)
	return L().With(fields...)
}

func getCaller(skip int) (name, file string, line int, ok bool) {
	pc, file, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return
	}
	name = runtime.FuncForPC(pc).Name()
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '/' {
			name = name[i+1:]
			break
		}
	}
	pathSepCnt := 0
	for i := len(file) - 1; i >= 0; i-- {
		if file[i] == '/' {
			pathSepCnt++
			if pathSepCnt == 2 {
				file = file[i+1:]
				break
			}
		}
	}
	return
}
