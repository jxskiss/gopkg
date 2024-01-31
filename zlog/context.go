package zlog

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ctxKey int

const (
	fieldsKey ctxKey = iota
	loggerKey
)

// CtxHandler customizes a logger's behavior in runtime dynamically.
type CtxHandler struct {

	/*
		// ChangeLevel returns a non-nil Level if it wants to change
		// the logger's logging level according to ctx.
		// It returns nil to keep the logger's logging level as-is.
		ChangeLevel func(ctx context.Context) *Level
	*/

	// WithCtx is called by Logger.Ctx, SugaredLogger.Ctx and
	// global functions WithCtx and SWithCtx to check ctx for additional
	// logging data.
	// It returns CtxResult to customize the logger's behavior.
	WithCtx func(ctx context.Context) CtxResult
}

// CtxResult holds information get from a context.
type CtxResult struct {
	// Non-nil Level changes the logger's logging level.
	Level *Level

	// Fields will be added to the logger as additional fields.
	Fields []zap.Field
}

// AddFields add logging fields to ctx which can be retrieved by
// GetFields, WithCtx, SWithCtx.
// Duplicate fields override the old ones in ctx.
func AddFields(ctx context.Context, fields ...zap.Field) context.Context {
	if len(fields) == 0 {
		return ctx
	}
	old, ok := ctx.Value(fieldsKey).([]zap.Field)
	if ok {
		fields = appendFields(old, fields)
	}
	return context.WithValue(ctx, fieldsKey, fields)
}

// GetFields returns the logging fields added to ctx by AddFields.
func GetFields(ctx context.Context) []zap.Field {
	fs, ok := ctx.Value(fieldsKey).([]zap.Field)
	if ok {
		return fs[:len(fs):len(fs)] // clip
	}
	return nil
}

// WithLogger returns a new context.Context with logger attached,
// which can be retrieved by WithCtx, SWithCtx.
func WithLogger[T Logger | SugaredLogger | *zap.Logger | *zap.SugaredLogger](
	ctx context.Context, logger T) context.Context {
	ctx = context.WithValue(ctx, loggerKey, logger)
	return ctx
}

func getLoggerFromCtx(ctx context.Context) any {
	return ctx.Value(loggerKey)
}

// Ctx creates a child logger, it calls GlobalConfig.CtxFunc to get CtxResult
// from ctx, it adds CtxResult.Fields to the child logger and changes
// the logger's level to CtxResult.Level, if it is not nil.
func (l Logger) Ctx(ctx context.Context, extra ...zap.Field) Logger {
	if ctx == nil {
		return l.With(extra...)
	}
	var fields []zap.Field
	logger := l
	ctxFunc := globals.Props.cfg.CtxHandler.WithCtx
	if ctxFunc != nil {
		ctxResult := ctxFunc(ctx)
		fields = ctxResult.Fields
		if ctxResult.Level != nil {
			logger = logger.WithOptions(zap.WrapCore(changeLevel(*ctxResult.Level)))
		}
	}
	fields = appendFields(fields, GetFields(ctx))
	fields = appendFields(fields, extra)
	if len(fields) > 0 {
		logger = logger.With(fields...)
	}
	return logger
}

// Ctx creates a child logger, it calls GlobalConfig.CtxFunc to get CtxResult
// from ctx, it adds CtxResult.Fields to the child logger and changes
// the logger's level to CtxResult.Level, if it is not nil.
func (s SugaredLogger) Ctx(ctx context.Context, extra ...zap.Field) SugaredLogger {
	if ctx == nil {
		return s.WithOptions(zap.Fields(extra...))
	}
	var fields []zap.Field
	logger := s
	ctxFunc := globals.Props.cfg.CtxHandler.WithCtx
	if ctxFunc != nil {
		ctxResult := ctxFunc(ctx)
		fields = ctxResult.Fields
		if ctxResult.Level != nil {
			logger = logger.WithOptions(zap.WrapCore(changeLevel(*ctxResult.Level)))
		}
	}
	fields = appendFields(fields, GetFields(ctx))
	fields = appendFields(fields, extra)
	if len(fields) > 0 {
		logger = logger.WithOptions(zap.Fields(fields...))
	}
	return logger
}

// WithCtx creates a child logger and customizes its behavior using context
// data (e.g. adding fields, dynamically changing level, etc.)
//
// If ctx is created by WithLogger, it carries a logger instance,
// this function uses that logger as a base to create the child logger,
// else it calls Logger.Ctx to build the child logger with contextual fields
// and optional dynamic level from ctx.
//
// Also see WithLogger, GlobalConfig.CtxFunc, CtxArgs and CtxResult
// for more details.
func WithCtx(ctx context.Context, extra ...zap.Field) Logger {
	if ctx == nil {
		return WithFields(extra...)
	}
	if lg := getLoggerFromCtx(ctx); lg != nil {
		var logger Logger
		switch x := lg.(type) {
		case Logger:
			logger = x
		case SugaredLogger:
			logger = x.Desugar()
		case *zap.Logger:
			logger = Logger{Logger: x}
		case *zap.SugaredLogger:
			logger = Logger{Logger: x.Desugar()}
		}
		if logger.Logger != nil {
			return logger.With(extra...)
		}
	}
	return L().Ctx(ctx, extra...)
}

// SWithCtx creates a child logger and customizes its behavior using context
// data (e.g. adding fields, dynamically changing level, etc.)
//
// If ctx is created by WithLogger, it carries a logger instance,
// this function uses that logger as a base to create the child logger,
// else it calls SugaredLogger.Ctx to build the child logger with
// contextual fields and optional dynamic level from ctx.
//
// Also see WithLogger, GlobalConfig.CtxFunc, CtxArgs and CtxResult
// for more details.
func SWithCtx(ctx context.Context, extra ...zap.Field) SugaredLogger {
	if ctx == nil {
		return S().WithOptions(zap.Fields(extra...))
	}
	if lg := getLoggerFromCtx(ctx); lg != nil {
		var logger SugaredLogger
		switch x := lg.(type) {
		case Logger:
			logger = x.Sugar()
		case SugaredLogger:
			logger = x
		case *zap.Logger:
			logger = SugaredLogger{SugaredLogger: x.Sugar()}
		case *zap.SugaredLogger:
			logger = SugaredLogger{SugaredLogger: x}
		}
		if logger.SugaredLogger != nil {
			if len(extra) > 0 {
				logger = logger.WithOptions(zap.Fields(extra...))
			}
			return logger
		}
	}
	return S().Ctx(ctx, extra...)
}

//nolint:predeclared
func appendFields(old []zap.Field, new []zap.Field) []zap.Field {
	if len(new) == 0 {
		return old
	}
	result := make([]zap.Field, len(old), len(old)+len(new))
	copy(result, old)

	// check namespace
	nsIdx := 0
	for i := len(result) - 1; i >= 0; i-- {
		if result[i].Type == zapcore.NamespaceType {
			nsIdx = i + 1
			break
		}
	}

	var hasNewNamespace bool
loop:
	for _, f := range new {
		if !hasNewNamespace {
			if f.Type == zapcore.NamespaceType {
				hasNewNamespace = true
				result = append(result, f)
				continue loop
			}
			for i := nsIdx; i < len(result); i++ {
				if result[i].Key == f.Key {
					result[i] = f
					continue loop
				}
			}
		}
		result = append(result, f)
	}
	return result
}
