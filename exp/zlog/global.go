package zlog

import (
	"context"
	"fmt"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	gL, gL_1 *zap.Logger
	gS, gS_1 *zap.SugaredLogger
	gP       *Properties
)

func init() {
	ReplaceGlobals(MustNewLogger(&Config{}))
}

// Properties records some information about zap.
type Properties struct {
	cfg    *Config
	level  atomicLevel
	core   zapcore.Core
	syncer zapcore.WriteSyncer
}

// ReplaceGlobals replaces the global Logger and SugaredLogger.
func ReplaceGlobals(logger *zap.Logger, props *Properties) {
	gL = logger
	gS = logger.Sugar()
	gP = props

	gL_1 = logger.WithOptions(zap.AddCallerSkip(1))
	gS_1 = logger.WithOptions(zap.AddCallerSkip(1)).Sugar()

	zap.ReplaceGlobals(logger)
}

// RedirectStdLog redirects output from the standard library's package-global
// logger to the global logger in this package. It returns a function to
// restore the original behavior of the standard library.
func RedirectStdLog() func() { return zap.RedirectStdLog(L()) }

// MustNewLogger initializes a zap logger, if error occurs it panics.
func MustNewLogger(cfg *Config, opts ...zap.Option) (*zap.Logger, *Properties) {
	logger, props, err := NewLogger(cfg, opts...)
	if err != nil {
		panic(fmt.Sprintf("invalid config to initialize logger: %v", err))
	}
	return logger, props
}

// NewLogger initializes a zap logger.
func NewLogger(cfg *Config, opts ...zap.Option) (*zap.Logger, *Properties, error) {
	cfg.fillDefaults()
	var output zapcore.WriteSyncer
	if len(cfg.File.Filename) > 0 {
		out, err := cfg.buildFileLogger()
		if err != nil {
			return nil, nil, err
		}
		output = zapcore.AddSync(out)
	} else {
		stderr, _, err := zap.Open("stderr")
		if err != nil {
			return nil, nil, err
		}
		output = stderr
	}
	return NewLoggerWithSyncer(cfg, output, opts...)
}

// NewLoggerWithSyncer initializes a zap logger with given write syncer.
func NewLoggerWithSyncer(cfg *Config, output zapcore.WriteSyncer, opts ...zap.Option) (*zap.Logger, *Properties, error) {
	cfg.fillDefaults()
	level := newAtomicLevel()
	err := level.UnmarshalText([]byte(cfg.Level))
	if err != nil {
		return nil, nil, err
	}
	encoder, err := cfg.buildEncoder()
	if err != nil {
		return nil, nil, err
	}
	core := zapcore.NewCore(encoder, output, level.zl)
	opts = append(cfg.buildOptions(output), opts...)
	lg := zap.New(core, opts...)
	prop := &Properties{
		cfg:    cfg,
		level:  level,
		core:   core,
		syncer: output,
	}
	return lg, prop, nil
}

// GetLevel gets the global logging level.
func GetLevel() Level { return gP.level.Level() }

// SetLevel alters the global logging level.
func SetLevel(lvl Level) { gP.level.SetLevel(lvl) }

// L returns the global Logger, which can be reconfigured with
// ReplaceGlobals.
func L() *zap.Logger { return gL }

// S returns the global SugaredLogger, which can be reconfigured with
// ReplaceGlobals.
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

func Debugf(format string, args ...interface{})  { _s().Debugf(format, args...) }
func Infof(format string, args ...interface{})   { _s().Infof(format, args...) }
func Warnf(format string, args ...interface{})   { _s().Warnf(format, args...) }
func Errorf(format string, args ...interface{})  { _s().Errorf(format, args...) }
func DPanicf(format string, args ...interface{}) { _s().DPanicf(format, args...) }
func Panicf(format string, args ...interface{})  { _s().Panicf(format, args...) }
func Fatalf(format string, args ...interface{})  { _s().Fatalf(format, args...) }

// With creates a child logger and adds structured context to it.
// Fields added to the child don't affect the parent, and vice versa.
func With(fields ...zap.Field) *zap.Logger {
	return L().With(fields...)
}

// WithCtx creates a child logger and adds fields extracted from ctx and extra.
//
// Note: to use this, Config.CtxFunc must be set, else it logs an error
// message at DPANIC level.
func WithCtx(ctx context.Context, extra ...zap.Field) *zap.Logger {
	ctxFunc := gP.cfg.CtxFunc
	if ctxFunc == nil {
		L().DPanic("calling WithCtx without CtxFunc configured")
		return L().With(extra...)
	}
	if ctx == nil {
		return L().With(extra...)
	}
	ctxFields := ctxFunc(ctx)
	if len(ctxFields) == 0 {
		return L().With(extra...)
	}
	if len(extra) == 0 {
		return L().With(ctxFields...)
	}
	return L().With(append(ctxFields, extra...)...)
}

// WithMethod creates a child logger and adds the caller's method name and
// extra fields.
func WithMethod(extra ...zap.Field) *zap.Logger {
	if gP.cfg.FunctionKey != "" {
		return L().With(extra...)
	}
	methodName, ok := getFunctionName(1)
	if !ok {
		return L().With(extra...)
	}
	if len(extra) == 0 {
		return L().With(zap.String(MethodKey, methodName))
	}
	fields := append([]zap.Field{zap.String(MethodKey, methodName)}, extra...)
	return L().With(fields...)
}

func getFunctionName(skip int) (name string, ok bool) {
	const skipOffset = 2 // skip getFunctionName and Callers
	pc := make([]uintptr, 1)
	numFrames := runtime.Callers(skip+skipOffset, pc)
	if numFrames < 1 {
		return
	}
	frame, _ := runtime.CallersFrames(pc).Next()
	return frame.Function, frame.PC != 0
}
