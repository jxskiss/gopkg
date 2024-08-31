// Package zlog provides opinionated high-level logging facilities
// based on go.uber.org/zap.
//
// Logger and SugaredLogger are simple wrappers of *zap.Logger and *zap.SugaredLogger,
// to provide more user-friendly API in addition to zap.
//
// # TraceLevel
//
// This package defines an extra [TraceLevel], for the most fine-grained
// information which helps developer "tracing" their code.
// Users can expect this logging level to be very verbose, you can use it,
// for example, to annotate each step in an algorithm or each individual
// calling with parameters in your program.
//
// In production deployment, it's strongly suggested to disable TraceLevel logs.
// In development, user may also configure to only allow or deny TraceLevel messages
// from some packages or some files to reduce the number of tracing logs.
// See [Logger.Trace], [SugaredLogger.Tracef], [TRACE], [GlobalConfig].TraceFilterRule
// for detailed documents.
//
// # Dynamic Level
//
// This package supports changing logging level dynamically, in two ways:
//
//  1. Loggers creates by this package wraps the zap Core by a dynamic-level Core.
//     User can use [Config].PerLoggerLevels to configure different level for different
//     loggers by logger names.
//     The format is "loggerName.subLogger=level".
//     If a level is configured for a parent logger, bug not configured for a child logger,
//     the child logger derives from its parent.
//  2. [GlobalConfig].CtxHandler, if configured, can optionally change level according to
//     contextual information, by returning a non-nil [CtxResult].Level value.
//
// # Context Integration
//
// This package integrates with [context.Context], user may add contextual fields
// to a Context (see [AddFields], [CtxHandler]), or add a pre-built logger
// to a Context (see [WithLogger]), or set a Context to use dynamic level
// (see [CtxHandler]), either smaller or greater than the base logger.
// Functions [Logger.Ctx], [SugaredLogger.Ctx] and [WithCtx]
// create child loggers with contextual information retrieved from a Context.
//
// # Multi-files Support
//
// [Config].PerLoggerFiles optionally set different file destination for different
// loggers specified by logger name.
// If a destination is configured for a parent logger, but not configured
// for a child logger, the child logger derives from its parent.
//
// # "logfmt" Encoder
//
// NewLogfmtEncoder creates a [zapcore.Encoder] which encodes log in the "logfmt" format.
// The returned encoder does not support [zapcore.ObjectMarshaler].
//
// # "logr" Adapter
//
// This package provides an adapter implementation of [logr.LogSink]
// to send logs from a [logr.Logger] to an underlying [zap.Logger].
//
// NewLogrLogger accepts optional options and creates a new logr.Logger
// using a [zap.Logger] as the underlying LogSink.
//
// # "slog" Adapter
//
// This package provides an adapter implementation of [slog.Handler]
// to send logs from a [slog.Logger] to an underlying [zap.Logger].
//
// NewSlogLogger accepts optional options and creates a new slog.Logger
// using a [zap.Logger] as the underlying Handler.
//
// # Example
//
//	func main() {
//		// Simple shortcut for development.
//		zlog.SetDevelopment()
//
//	 	// Or, provide complete config and options to configure it.
//		logCfg := &zlog.Config{ /* ... */ }
//		logOpts := []zap.Option{ /* ... */ }
//		zlog.SetupGlobals(logCfg, logOpts...)
//
//		// Use the loggers ...
//		zlog.L() /* ... */
//		zlog.S() /* ... */
//		zlog.With( ... ) /* ... */
//		zlog.SugaredWith( ... ) /* ... */
//
//		// Use with context integration.
//		ctx = zlog.AddFields(ctx, ... ) // or ctx = zlog.WithLogger( ... )
//		zlog.L().Ctx(ctx) /* ... */
//		zlog.S().Ctx(ctx) /* ... */
//		zlog.WithCtx(ctx) /* ... */
//
//		// logr
//		logger := zlog.NewLogrLogger( /* ... */ )
//
//		// slog
//		logger := zlog.NewSlogLogger( /* ... */ )
//
//		// ......
//	}
package zlog

import (
	"runtime"

	"go.uber.org/zap"
)

// Logger is a simple wrapper of *zap.Logger.
// It provides fast, leveled, structured logging. All methods are safe
// for concurrent use.
// A zero Logger is not ready to use, don't construct Logger instances
// outside this package.
type Logger struct {
	*zap.Logger

	// Note, this type is designed to be a simple wrapper,
	// don't add more fields to it.
	_ struct{}
}

// Named adds a new path segment to the logger's name. Segments are joined by
// periods. By default, Loggers are unnamed.
func (l Logger) Named(name string) Logger {
	return Logger{Logger: l.Logger.Named(name)}
}

// With creates a child logger and adds structured context to it.
// Fields added to the child don't affect the parent, and vice versa.
// Any fields that require evaluation (such as Objects) are evaluated
// upon invocation of With.
func (l Logger) With(fields ...zap.Field) Logger {
	return Logger{Logger: l.Logger.With(fields...)}
}

// WithLazy creates a child logger and adds structured context to it lazily.
//
// The fields are evaluated only if the logger is further chained with [With]
// or is written to with any of the log level methods.
// Until that occurs, the logger may retain references to objects inside the fields,
// and logging will reflect the state of an object at the time of logging,
// not the time of WithLazy().
//
// WithLazy provides a worthwhile performance optimization for contextual loggers
// when the likelihood of using the child logger is low,
// such as error paths and rarely taken branches.
//
// Similar to [With], fields added to the child don't affect the parent, and vice versa.
func (l Logger) WithLazy(fields ...zap.Field) Logger {
	return Logger{Logger: l.Logger.WithLazy(fields...)}
}

// WithMethod adds the caller's method name as a context field,
// using the globally configured GlobalConfig.MethodNameKey as key.
func (l Logger) WithMethod() Logger {
	methodName, _, _, _, ok := getCaller(1)
	if !ok {
		return l
	}
	methodNameKey := globals.Props.cfg.MethodNameKey
	return l.With(zap.String(methodNameKey, methodName))
}

// WithOptions clones the current Logger, applies the supplied Options,
// and returns the resulting Logger. It's safe to use concurrently.
func (l Logger) WithOptions(opts ...zap.Option) Logger {
	return Logger{Logger: l.Logger.WithOptions(opts...)}
}

// Sugar clones the current Logger, returns a SugaredLogger.
//
// Sugaring a Logger is quite inexpensive, so it's reasonable for a
// single application to use both Loggers and SugaredLoggers, converting
// between them on the boundaries of performance-sensitive code.
func (l Logger) Sugar() SugaredLogger {
	return SugaredLogger{SugaredLogger: l.Logger.Sugar()}
}

// SugaredLogger is a simple wrapper of *zap.SugaredLogger.
// It provides a slower, but less verbose, API.
// A zero SugaredLogger is not ready to use, don't construct
// SugaredLogger instances outside this package.
//
// Any Logger can be converted to a SugaredLogger with its Sugar method.
type SugaredLogger struct {
	*zap.SugaredLogger

	// Note, this type is designed to be a simple wrapper,
	// don't add more fields to it.
	_ struct{}
}

// Named adds a sub-scope to the logger's name. See Logger.Named for details.
func (s SugaredLogger) Named(name string) SugaredLogger {
	return SugaredLogger{SugaredLogger: s.SugaredLogger.Named(name)}
}

// With adds a variadic number of fields to the logging context.
// It accepts a mix of strongly-typed Field objects and loosely-typed
// key-value pairs. When processing pairs, the first element of the pair
// is used as the field key and the second as the field value.
//
// Note that the keys in key-value pairs should be strings.
// In development, passing a non-string key panics.
// In production, the logger is more forgiving: a separate error is logged,
// but the key-value pair is skipped and execution continues.
// Passing an orphaned key triggers similar behavior:
// panics in development and errors in production.
func (s SugaredLogger) With(args ...any) SugaredLogger {
	return SugaredLogger{SugaredLogger: s.SugaredLogger.With(args...)}
}

// WithMethod adds the caller's method name as a context field,
// using the globally configured GlobalConfig.MethodNameKey as key.
func (s SugaredLogger) WithMethod() SugaredLogger {
	methodName, _, _, _, ok := getCaller(1)
	if !ok {
		return s
	}
	methodNameKey := globals.Props.cfg.MethodNameKey
	return s.With(zap.String(methodNameKey, methodName))
}

// WithOptions clones the current SugaredLogger, applies the supplied Options,
// and returns the result. It's safe to use concurrently.
func (s SugaredLogger) WithOptions(opts ...zap.Option) SugaredLogger {
	return SugaredLogger{SugaredLogger: s.SugaredLogger.WithOptions(opts...)}
}

// Desugar unwraps a SugaredLogger, returning the original Logger.
//
// Desugaring is quite inexpensive, so it's reasonable for a single
// application to use both Loggers and SugaredLoggers, converting
// between them on the boundaries of performance-sensitive code.
func (s SugaredLogger) Desugar() Logger {
	return Logger{Logger: s.SugaredLogger.Desugar()}
}

func getCaller(skip int) (funcName, fullFileName, simpleFileName string, line int, ok bool) {
	pc, fullFileName, line, ok := runtime.Caller(skip + 1)
	if !ok {
		return
	}
	fs := runtime.CallersFrames([]uintptr{pc})
	frame, _ := fs.Next()
	funcName = frame.Func.Name()
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

// With creates a child logger and adds structured context to it.
// Fields added to the child don't affect the parent, and vice versa.
func With(fields ...zap.Field) Logger {
	return L().With(fields...)
}

// SugaredWith creates a child logger and adds a variadic number of fields
// to the logging context.
// Fields added to the child don't affect the parent, and vice versa.
//
// It accepts a mix of strongly-typed Field objects and loosely-typed
// key-value pairs. When processing pairs, the first element of the pair
// is used as the field key and the second as the field value.
func SugaredWith(args ...any) SugaredLogger {
	return S().With(args...)
}
