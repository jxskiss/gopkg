package zlog

import (
	"context"
	"unsafe"

	"go.uber.org/zap"
)

type ctxKey int

const (
	fieldsKey ctxKey = iota
	loggerKey
)

// AddFields add logging fields to ctx which can be retrieved by GetFields.
// Duplicate field overrides the old in ctx.
func AddFields(ctx context.Context, fields ...zap.Field) context.Context {
	if len(fields) == 0 {
		return ctx
	}
	var fs []zap.Field
	old, ok := ctx.Value(fieldsKey).([]zap.Field)
	if ok {
		fs = make([]zap.Field, len(old), len(old)+len(fields))
		copy(fs, old)
	NEXT:
		for i := range fields {
			key := fields[i].Key
			for j := range fs {
				if fs[j].Key == key {
					fs[j] = fields[i]
					continue NEXT
				}
			}
			fs = append(fs, fields[i])
		}
	} else {
		fs = fields
	}
	return context.WithValue(ctx, fieldsKey, fs)
}

// GetFields returns the logging fields associated with ctx.
func GetFields(ctx context.Context) []zap.Field {
	fs, ok := ctx.Value(fieldsKey).([]zap.Field)
	if ok {
		return fs[:len(fs):len(fs)] // clip
	}
	return nil
}

// WithLogger returns a new context.Context with logger attached,
// which can be retrieved by calling GetLogger.
func WithLogger[T zap.Logger | zap.SugaredLogger](ctx context.Context, logger *T) context.Context {
	if unsafe.Pointer(logger) != nil {
		ctx = context.WithValue(ctx, loggerKey, logger)
	}
	return ctx
}

// GetLogger returns the logger associated with ctx.
// If there is no logger associated with ctx, it checks for associated
// logging fields and returns a new *zap.Logger with the fields.
// In case that no fields available, it returns a basic *zap.Logger.
// Fields specified by param extra will be added to the returned logger.
func GetLogger(ctx context.Context, extra ...zap.Field) *zap.Logger {
	if lg := ctx.Value(loggerKey); lg != nil {
		switch x := lg.(type) {
		case *zap.Logger:
			return x.With(extra...)
		case *zap.SugaredLogger:
			return x.Desugar().With(extra...)
		}
	}
	fs := GetFields(ctx)
	if len(extra) == 0 {
		return L().With(fs...)
	}
	if len(fs) == 0 {
		return L().With(extra...)
	}
	return L().With(append(fs, extra...)...)
}

// GetSugaredLogger returns the logger associated with ctx.
// If there is no logger associated with ctx, it checks for associated
// logger fields and returns a new *zap.SugaredLogger with the fields.
// In case that no fields available, it returns a basic *zap.SugaredLogger.
// Fields specified by param extra will be added to the returned logger.
func GetSugaredLogger(ctx context.Context, extra ...any) *zap.SugaredLogger {
	if lg := ctx.Value(loggerKey); lg != nil {
		switch x := lg.(type) {
		case *zap.Logger:
			return x.Sugar().With(extra...)
		case *zap.SugaredLogger:
			return x.With(extra...)
		}
	}
	fs := GetFields(ctx)
	if len(extra) == 0 {
		if len(fs) == 0 {
			return S()
		}
		return L().With(fs...).Sugar()
	}
	if len(fs) == 0 {
		if len(extra) == 0 {
			return S()
		}
		return S().With(extra...)
	}
	mergeFields := make([]any, len(fs)+len(extra))
	for i := range fs {
		mergeFields[i] = fs[i]
	}
	copy(mergeFields[len(fs):], extra)
	return S().With(mergeFields...)
}
