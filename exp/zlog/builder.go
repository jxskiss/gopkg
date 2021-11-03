package zlog

import (
	"context"
	"sync/atomic"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var baseBuilder = &Builder{}

type ctxBuilderKey struct{}

func getCtxBuilder(ctx context.Context) *Builder {
	builder, _ := ctx.Value(ctxBuilderKey{}).(*Builder)
	return builder
}

// B returns a Builder with given ctx.
//
// If ctx is nil, it returns an empty Builder, else if the ctx is created
// by WithBuilder, then it carries a Builder instance, this function
// returns that Builder.
// Otherwise, if the ctx is not nil and GlobalConfig.CtxFunc is configured
// globally, it calls the CtxFunc to get CtxResult from ctx.
func B(ctx context.Context) *Builder {
	builder := baseBuilder
	if ctx != nil {
		if ctxBuilder := getCtxBuilder(ctx); ctxBuilder != nil {
			builder = ctxBuilder
		} else if ctxFunc := gP.cfg.CtxFunc; ctxFunc != nil {
			builder = builder.Ctx(ctx)
		}
	}
	return builder
}

// WithBuilder returns a copy of parent ctx with builder associated.
// The associated builder can be accessed by B or WithCtx.
func WithBuilder(ctx context.Context, builder *Builder) context.Context {
	return context.WithValue(ctx, ctxBuilderKey{}, builder)
}

// Builder provides chaining methods to build a logger.
// Different with calling zap.Logger's methods, it does not write the context
// information to underlying buffer immediately, later data will override
// the former which has same key within a namespace. When the builder is
// prepared, call Build to get the final logger.
//
// A Builder is safe for concurrent use, it will copy data if necessary.
// User may pass a Builder across functions to handle duplicate.
//
// A zero value for Builder is ready to use. A Builder must not be copied
// after first use.
type Builder struct {
	base       *zap.Logger
	level      *Level
	fields     []zap.Field
	name       string
	methodName string

	final atomic.Value
}

func (b *Builder) clone() *Builder {
	n := len(b.fields)
	return &Builder{
		base:       b.base,
		level:      b.level,
		fields:     b.fields[:n:n],
		name:       b.name,
		methodName: b.methodName,
	}
}

func (b *Builder) getBaseLogger() *zap.Logger {
	if b.base != nil {
		return b.base
	}
	return L()
}

// Base sets the base logger to build upon.
func (b *Builder) Base(logger *zap.Logger) *Builder {
	if logger == nil {
		return b
	}
	out := b.clone()
	out.base = logger
	return out
}

// Ctx customizes the logger's behavior using context data (e.g. adding
// fields, dynamically change logging level, etc.)
// See GlobalConfig.CtxFunc, CtxArgs and CtxResult for details.
//
// It calls GlobalConfig.CtxFunc to get CtxResult from ctx, in case that
// GlobalConfig.CtxFunc is not configured globally, it logs an error
// message at DPANIC level.
func (b *Builder) Ctx(ctx context.Context) *Builder {
	if ctx == nil {
		return b
	}
	ctxFunc := gP.cfg.CtxFunc
	if ctxFunc == nil {
		L().DPanic("calling Builder.Ctx without CtxFunc configured")
		return b
	}
	ctxResult := ctxFunc(ctx, CtxArgs{})
	return b.withCtxResult(ctxResult)
}

func (b *Builder) withCtxResult(ctxResult CtxResult) *Builder {
	if len(ctxResult.Fields) == 0 && ctxResult.Level == nil {
		return b
	}
	out := b.clone()
	out.fields = appendFields(out.fields, ctxResult.Fields)
	if ctxResult.Level != nil {
		out.level = ctxResult.Level
	}
	return out
}

// With adds extra fields to the builder.
// Duplicate keys override the old ones within a namespace.
func (b *Builder) With(fields ...zap.Field) *Builder {
	if len(fields) == 0 {
		return b
	}
	out := b.clone()
	out.fields = appendFields(out.fields, fields)
	return out
}

// Method adds the caller's method name to the builder.
func (b *Builder) Method() *Builder {
	out := b.clone()
	out.methodName, _, _, _ = getCaller(1)
	return out
}

// Named adds a new path segment to the logger's name. By default,
// loggers are unnamed.
func (b *Builder) Named(name string) *Builder {
	out := b.clone()
	out.name = name
	return out
}

// Level optionally changes the level of a logger. By default,
// a child logger has same level with its parent.
func (b *Builder) Level(level Level) *Builder {
	out := b.clone()
	out.level = &level
	return out
}

func (b *Builder) getFinalLogger() *zap.Logger {
	if final := b.final.Load(); final != nil {
		return final.(*zap.Logger)
	}
	final := b.getBaseLogger()
	if b.name != "" {
		final = final.Named(b.name)
	}
	if b.methodName == "" {
		final = final.With(b.fields...)
	} else {
		methodNameKey := gP.cfg.MethodNameKey
		fields := append([]zap.Field{zap.String(methodNameKey, b.methodName)}, b.fields...)
		final = final.With(fields...)
	}
	if b.level != nil {
		final = final.WithOptions(zap.WrapCore(tryChangeLevel(*b.level)))
	}
	b.final.Store(final)
	return final
}

// Build builds and returns the final logger.
func (b *Builder) Build() *zap.Logger {
	return b.getFinalLogger()
}

// Sugar builds and returns the final sugared logger.
// It's a shortcut for Builder.Build().Sugar()
func (b *Builder) Sugar() *zap.SugaredLogger {
	return b.getFinalLogger().Sugar()
}

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
