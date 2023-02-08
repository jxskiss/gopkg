package zlog

import (
	"context"
	"fmt"
	"runtime"

	"go.uber.org/zap"
)

var (
	gL, gL_1 *zap.Logger
	gS, gS_1 *zap.SugaredLogger
	gP       *Properties
)

func init() {
	ReplaceGlobals(mustNewGlobalLogger(&Config{}))
}

// Properties records some information about the global config.
type Properties struct {
	cfg   GlobalConfig
	level atomicLevel
}

func (p *Properties) setup() func() {
	if p.cfg.MethodNameKey == "" {
		p.cfg.MethodNameKey = defaultMethodNameKey
	}
	var resetStdLog = func() {}
	if p.cfg.RedirectStdLog {
		resetStdLog = RedirectStdLog(L())
	}
	oldDisableTrace := disableTrace
	var resetDisableTrace = func() {
		disableTrace = oldDisableTrace
	}
	disableTrace = p.cfg.DisableTrace
	return func() {
		resetDisableTrace()
		resetStdLog()
	}
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
// See Config and GlobalConfig for available configurations.
//
// It should be called at program startup, library code shall not touch
// this function.
func SetupGlobals(cfg *Config, opts ...zap.Option) {
	ReplaceGlobals(mustNewGlobalLogger(cfg, opts...))
}

func mustNewGlobalLogger(cfg *Config, opts ...zap.Option) (*zap.Logger, *Properties) {
	logger, props, err := New(cfg, opts...)
	if err != nil {
		panic(fmt.Sprintf("zlog: invalid config to initialize logger: %v", err))
	}
	return logger, props
}

// ReplaceGlobals replaces the global Logger and SugaredLogger, and returns a
// function to restore the original values.
//
// It should be called at program startup, library code shall not touch
// this function.
func ReplaceGlobals(logger *zap.Logger, props *Properties) func() {
	oldL, oldP := gL, gP

	gL = logger
	gS = logger.Sugar()
	gP = props

	gL_1 = logger.WithOptions(zap.AddCallerSkip(1))
	gS_1 = gL_1.Sugar()

	resetProps := props.setup()
	zap.ReplaceGlobals(logger)

	return func() {
		resetProps()
		ReplaceGlobals(oldL, oldP)
	}
}

// SetDevelopment sets the global logger in development mode, and redirects
// output from the standard log library's package-global logger to the
// global logger in this package.
//
// It should only be called at program startup, when you run in development
// mode, for production mode, please check SetupGlobals and ReplaceGlobals.
func SetDevelopment() {
	cfg := &Config{}
	cfg.Development = true
	cfg.RedirectStdLog = true
	ReplaceGlobals(mustNewGlobalLogger(cfg))
}

// GetLevel gets the global logging level.
func GetLevel() Level { return gP.GetLevel() }

// SetLevel modifies the global logging level on the fly.
// It's safe for concurrent use.
func SetLevel(lvl Level) { gP.SetLevel(lvl) }

// L returns the global Logger, which can be reconfigured with
// SetupGlobals and ReplaceGlobals.
func L() *zap.Logger { return gL }

// S returns the global SugaredLogger, which can be reconfigured with
// SetupGlobals and ReplaceGlobals.
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

func Debug(msg string, fields ...zap.Field)  { _l().Debug(msg, fields...) }
func Info(msg string, fields ...zap.Field)   { _l().Info(msg, fields...) }
func Warn(msg string, fields ...zap.Field)   { _l().Warn(msg, fields...) }
func Error(msg string, fields ...zap.Field)  { _l().Error(msg, fields...) }
func DPanic(msg string, fields ...zap.Field) { _l().DPanic(msg, fields...) }
func Panic(msg string, fields ...zap.Field)  { _l().Panic(msg, fields...) }
func Fatal(msg string, fields ...zap.Field)  { _l().Fatal(msg, fields...) }

func Debugf(format string, args ...any)  { _s().Debugf(format, args...) }
func Infof(format string, args ...any)   { _s().Infof(format, args...) }
func Warnf(format string, args ...any)   { _s().Warnf(format, args...) }
func Errorf(format string, args ...any)  { _s().Errorf(format, args...) }
func DPanicf(format string, args ...any) { _s().DPanicf(format, args...) }
func Panicf(format string, args ...any)  { _s().Panicf(format, args...) }
func Fatalf(format string, args ...any)  { _s().Fatalf(format, args...) }

// Print uses fmt.Sprint to log a message at InfoLevel if it's enabled.
//
// It has same signature with log.Print, which helps to migrate from the
// standard library to this package.
func Print(args ...any) {
	if l := _l(); l.Core().Enabled(zap.InfoLevel) {
		l.Info(fmt.Sprint(args...))
	}
}

// Printf logs a message at InfoLevel if it's enabled.
//
// It has same signature with log.Printf, which helps to migrate from the
// standard library to this package.
func Printf(format string, args ...any) {
	if l := _l(); l.Core().Enabled(zap.InfoLevel) {
		l.Info(fmt.Sprintf(format, args...))
	}
}

// Println logs a message at InfoLevel if it's enabled.
//
// It has same signature with log.Println, which helps to migrate from the
// standard library to this package.
func Println(args ...any) {
	if l := _l(); l.Core().Enabled(zap.InfoLevel) {
		l.Info(fmt.Sprintln(args...))
	}
}

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
// GlobalConfig.CtxFunc to get CtxResult from ctx. In case that
// GlobalConfig.CtxFunc is not configured globally, it logs an error
// message at DPANIC level.
//
// Also see GlobalConfig.CtxFunc, CtxArgs and CtxResult for more details.
func WithCtx(ctx context.Context, extra ...zap.Field) *zap.Logger {
	if ctx == nil {
		return L().With(extra...)
	}
	if builder := getCtxBuilder(ctx); builder != nil {
		return builder.With(extra...).Build()
	}
	ctxFunc := gP.cfg.CtxFunc
	if ctxFunc == nil {
		L().DPanic("calling WithCtx without CtxFunc configured")
		return L().With(extra...)
	}
	ctxResult := ctxFunc(ctx, CtxArgs{})
	return B(nil).withCtxResult(ctxResult).With(extra...).Build()
}

// WithMethod creates a child logger and adds the caller's method name
// to the logger.
// It also adds the given extra fields to the logger.
func WithMethod(extra ...zap.Field) *zap.Logger {
	methodName, _, _, ok := getCaller(1)
	if !ok {
		return L().With(extra...)
	}
	methodNameKey := gP.cfg.MethodNameKey
	if len(extra) == 0 {
		return L().With(zap.String(methodNameKey, methodName))
	}
	fields := append([]zap.Field{zap.String(methodNameKey, methodName)}, extra...)
	return L().With(fields...)
}

// Named creates a child logger and adds a new name segment to the logger's
// name. By default, loggers are unnamed.
// It also adds the given extra fields to the logger.
func Named(name string, extra ...zap.Field) *zap.Logger {
	return L().Named(name).With(extra...)
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
