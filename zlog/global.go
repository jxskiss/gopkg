package zlog

import (
	"sync/atomic"

	"go.uber.org/zap"

	"github.com/jxskiss/gopkg/v2/internal/logfilter"
)

var globals struct {
	Default, Skip1 struct {
		L *zap.Logger
		S *zap.SugaredLogger
	}
	Props *Properties

	// Level is a copy of Props.level for fast path accessing.
	Level atomic.Int32
}

func init() {
	ReplaceGlobals(mustNewGlobalLogger(&Config{}))
}

// Properties records some information about the global config.
type Properties struct {
	cfg         GlobalConfig
	level       zap.AtomicLevel
	traceFilter *logfilter.FileNameFilter
}

func (p *Properties) setupGlobals() func() {
	if p.cfg.MethodNameKey == "" {
		p.cfg.MethodNameKey = defaultMethodNameKey
	}
	var resetStdLog = func() {}
	if p.cfg.RedirectStdLog {
		resetStdLog = RedirectStdLog(L().Logger.Named("stdlog"))
	}
	p.compileTraceFilter()
	globals.Level.Store(int32(p.level.Level()))
	return func() {
		resetStdLog()
	}
}

// ReplaceGlobals replaces the global Logger and SugaredLogger,
// and returns a function to restore the original values.
//
// It is meant to be called at program startup, library code shall not call
// this function.
func ReplaceGlobals(logger *zap.Logger, props *Properties) func() {
	oldL, oldP := globals.Default.L, globals.Props

	globals.Default.L = logger
	globals.Default.S = logger.Sugar()
	globals.Skip1.L = logger.WithOptions(zap.AddCallerSkip(1))
	globals.Skip1.S = globals.Skip1.L.Sugar()
	globals.Props = props

	resetProps := props.setupGlobals()
	zap.ReplaceGlobals(logger)

	return func() {
		resetProps()
		ReplaceGlobals(oldL, oldP)
	}
}

// SetDevelopment is a shortcut of SetupGlobals with default configuration
// for development. It sets the global logger in development mode,
// and redirects output from the standard log library's package-global
// logger to the global logger in this package.
//
// It is meant to be called at program startup, when you run in development
// mode, for production mode, please check SetupGlobals and ReplaceGlobals.
func SetDevelopment() {
	cfg := &Config{}
	cfg.Development = true
	cfg.RedirectStdLog = true
	ReplaceGlobals(mustNewGlobalLogger(cfg))
}

// SetupGlobals setups the global loggers in this package and zap library.
// By default, global loggers are set with default configuration with info
// level and json format, you may use this function to change the default
// loggers.
//
// See Config and GlobalConfig for available configurations.
//
// It is meant to be called at program startup, library code shall not call
// this function.
func SetupGlobals(cfg *Config, opts ...zap.Option) {
	ReplaceGlobals(mustNewGlobalLogger(cfg, opts...))
}

func mustNewGlobalLogger(cfg *Config, opts ...zap.Option) (*zap.Logger, *Properties) {
	logger, props, err := New(cfg, opts...)
	if err != nil {
		panic("zlog: invalid config to initialize logger: " + err.Error())
	}
	return logger, props
}

// GetLevel gets the global logging level.
func GetLevel() Level { return globals.Props.level.Level() }

// SetLevel modifies the global logging level on the fly.
// It's safe for concurrent use.
func SetLevel(lvl Level) {
	globals.Props.level.SetLevel(lvl)
	globals.Level.Store(int32(lvl))
}

// L returns the global Logger, which can be reconfigured with
// SetupGlobals and ReplaceGlobals.
func L() Logger { return Logger{Logger: globals.Default.L} }

// S returns the global SugaredLogger, which can be reconfigured with
// SetupGlobals and ReplaceGlobals.
func S() SugaredLogger { return SugaredLogger{SugaredLogger: globals.Default.S} }

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

func _l() Logger        { return Logger{Logger: globals.Skip1.L} }
func _s() SugaredLogger { return SugaredLogger{SugaredLogger: globals.Skip1.S} }

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
func Print(args ...any) { _s().Info(args...) }

// Printf logs a message at InfoLevel if it's enabled.
//
// It has same signature with log.Printf, which helps to migrate from the
// standard library to this package.
func Printf(format string, args ...any) { _s().Infof(format, args...) }

// Println logs a message at InfoLevel if it's enabled.
//
// It has same signature with log.Println, which helps to migrate from the
// standard library to this package.
func Println(args ...any) { _s().Infoln(args...) }
