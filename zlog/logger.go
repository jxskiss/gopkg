package zlog

import (
	"context"
	"log/slog"
	"os"
	"runtime"
	"time"

	slogconsolehandler "github.com/jxskiss/slog-console-handler"
)

const ErrorKey = "error"

type Logger = slog.Logger

// Default returns the default Logger.
func Default() *Logger {
	return slog.Default()
}

// SetDefault makes l the default Logger.
// After this call, output from the log package's default Logger
// (as with [log.Print], etc.) will be logged at LevelInfo using l's Handler.
func SetDefault(l *Logger) {
	slog.SetDefault(l)
	RedirectStdLog(nil, defaultStdLogAttrs)
}

// SetDevelopment sets the default logger to use [slogconsolehandler.Default]
// as the underlying handler.
func SetDevelopment(level slog.Level) {
	slogconsolehandler.SetLevel(level)
	handler := NewHandler(slogconsolehandler.Default, nil)
	SetDefault(slog.New(handler))
}

func With(ctx context.Context, args ...any) *Logger {
	return FromCtx(ctx).With(args...)
}

func WithError(ctx context.Context, err error) *Logger {
	if err == nil {
		return FromCtx(ctx)
	}
	return FromCtx(ctx).With(slog.Any(ErrorKey, err))
}

func WithGroup(ctx context.Context, group string) *Logger {
	return FromCtx(ctx).WithGroup(group)
}

func Debug(ctx context.Context, msg string, args ...any) {
	_log(0, ctx, FromCtx(ctx), slog.LevelDebug, msg, args)
}

func Info(ctx context.Context, msg string, args ...any) {
	_log(0, ctx, FromCtx(ctx), slog.LevelInfo, msg, args)
}

func Warn(ctx context.Context, msg string, args ...any) {
	_log(0, ctx, FromCtx(ctx), slog.LevelWarn, msg, args)
}

func Error(ctx context.Context, msg string, args ...any) {
	_log(0, ctx, FromCtx(ctx), slog.LevelError, msg, args)
}

func Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	_log(0, ctx, FromCtx(ctx), level, msg, args)
}

func LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	_logAttrs(0, ctx, FromCtx(ctx), level, msg, attrs)
}

func LogSkip(ctx context.Context, skip int, level slog.Level, msg string, args ...any) {
	_log(skip, ctx, FromCtx(ctx), level, msg, args)
}

func LogAttrsSkip(ctx context.Context, skip int, level slog.Level, msg string, attrs ...slog.Attr) {
	_logAttrs(skip, ctx, FromCtx(ctx), level, msg, attrs)
}

// Fatal is equivalent to Error() followed by a call to os.Exit(1).
func Fatal(ctx context.Context, msg string, args ...any) {
	_log(0, ctx, FromCtx(ctx), slog.LevelError, msg, args)
	os.Exit(1)
}

// _log is the low-level logging method for methods that take ...any.
// It must always be called directly by an exported logging method
// or function, because it uses a fixed call depth to obtain the pc.
func _log(skip int, ctx context.Context, l *Logger, level slog.Level, msg string, args []any) {
	if ctx == nil {
		ctx = context.Background()
	}
	if !l.Enabled(ctx, level) {
		return
	}

	var pc uintptr
	var pcs [1]uintptr
	runtime.Callers(skip+3, pcs[:])
	pc = pcs[0]

	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)
	_ = l.Handler().Handle(ctx, r)
}

// _logAttrs is like log, but for methods that take ...Attr.
func _logAttrs(skip int, ctx context.Context, l *Logger, level slog.Level, msg string, attrs []slog.Attr) {
	if ctx == nil {
		ctx = context.Background()
	}
	if !l.Enabled(ctx, level) {
		return
	}

	var pc uintptr
	var pcs [1]uintptr
	runtime.Callers(skip+3, pcs[:])
	pc = pcs[0]

	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.AddAttrs(attrs...)
	_ = l.Handler().Handle(ctx, r)
}
