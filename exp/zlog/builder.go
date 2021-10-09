package zlog

import (
	"context"
	"sync/atomic"

	"go.uber.org/zap"
)

var baseBuilder = &Builder{}

type ctxBuilderKey struct{}

func getCtxBuilder(ctx context.Context) *Builder {
	builder, _ := ctx.Value(ctxBuilderKey{}).(*Builder)
	return builder
}

// B returns a Builder with given ctx.
// If the ctx is created by WithBuilder, then it carries a Builder instance,
// this function returns that Builder.
// If ctx is not nil and Config.CtxFunc is configured globally, it calls
// the CtxFunc to extract fields from ctx and adds the fields to the builder.
// If ctx is nil, it returns an empty Builder.
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
	fields     []zap.Field
	name       string
	methodName string

	final atomic.Value
}

func (b *Builder) clone() *Builder {
	n := len(b.fields)
	return &Builder{
		base:       b.base,
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

// Ctx adds fields extracted from ctx to the builder.
// It calls Config.CtxFunc to extract fields from ctx. In case Config.CtxFunc
// is not configured globally, it logs an error message at DPANIC level.
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
	return b.With(ctxResult.Fields...)
}

// With adds extra fields to the builder. Duplicate keys override
// the old ones within a namespace.
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
	if gP.cfg.FunctionKey != "" {
		return b
	}
	out := b.clone()
	out.methodName, _ = getFunctionName(1)
	return out
}

// Named adds a new path segment to the logger's name. By default,
// loggers are unnamed.
func (b *Builder) Named(name string) *Builder {
	out := b.clone()
	out.name = name
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
		fields := append([]zap.Field{zap.String(MethodKey, b.methodName)}, b.fields...)
		final = final.With(fields...)
	}
	b.final.Store(final)
	return final
}

// L builds and returns the final zap logger.
func (b *Builder) L() *zap.Logger {
	return b.getFinalLogger()
}

// S builds and returns the final zap sugared logger.
func (b *Builder) S() *zap.SugaredLogger {
	return b.getFinalLogger().Sugar()
}
