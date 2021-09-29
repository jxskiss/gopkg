package zlog

import (
	"context"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var baseBuilder = &Builder{}

// B returns an empty Builder.
func B() *Builder { return baseBuilder }

// Builder provides chaining methods to build a logger.
// Different with calling zap.Logger's methods, it does not write the context
// information to underlying buffer immediately, later data will override
// the former which has same key within a namespace. When the builder is
// prepared, call Build to get the final logger.
//
// A Builder is safe for concurrent use, it will copy data if necessary.
// User may pass a Builder across functions to handle duplicate.
//
// A zero value for Builder is ready to use.
type Builder struct {
	base       *zap.Logger
	fields     []zap.Field
	name       string
	methodName string

	mu    sync.Mutex
	final *zap.Logger
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

// Ctx adds fields extracted from ctx to the builder.
//
// If the ctx is created by WithBuilder, it carries a Builder instance,
// this function returns that Builder, else it uses Config.CtxFunc to
// extract fields from ctx. In case Config.CtxFunc is not configured,
// it logs an error message at DPANIC level.
func (b *Builder) Ctx(ctx context.Context) *Builder {
	if ctx == nil {
		return b
	}
	if builder := getCtxBuilder(ctx); builder != nil {
		return builder
	}
	ctxFunc := gP.cfg.CtxFunc
	if ctxFunc == nil {
		L().DPanic("calling Builder.Ctx without CtxFunc configured")
		return b
	}
	return b.With(ctxFunc(ctx)...)
}

// Logger sets the base logger to build upon.
func (b *Builder) Logger(logger *zap.Logger) *Builder {
	if logger == nil {
		return b
	}
	out := b.clone()
	out.base = logger
	return out
}

// With adds extra fields to the builder. Duplicate keys override
// the old ones within a namespace.
func (b *Builder) With(fields ...zap.Field) *Builder {
	if len(fields) == 0 {
		return b
	}
	out := b.clone()
	out.fields = make([]zap.Field, len(b.fields), len(b.fields)+len(fields))
	copy(out.fields, b.fields)

	// check namespace
	nsIdx := 0
	for i := len(out.fields) - 1; i >= 0; i-- {
		if out.fields[i].Type == zapcore.NamespaceType {
			nsIdx = i + 1
			break
		}
	}

loop:
	for _, f := range fields {
		for i := nsIdx; i < len(out.fields); i++ {
			if out.fields[i].Key == f.Key {
				out.fields[i] = f
				continue loop
			}
		}
		out.fields = append(out.fields, f)
	}
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

// Build builds and returns the final logger.
func (b *Builder) Build() *zap.Logger {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.final != nil {
		return b.final
	}
	final := b.base
	if final == nil {
		final = L()
	}
	if b.name != "" {
		final = final.Named(b.name)
	}
	if b.methodName == "" {
		final = final.With(b.fields...)
	} else {
		fields := append([]zap.Field{zap.String(MethodKey, b.methodName)}, b.fields...)
		final = final.With(fields...)
	}
	b.final = final
	return final
}

type ctxBuilderKey struct{}

func getCtxBuilder(ctx context.Context) *Builder {
	builder, _ := ctx.Value(ctxBuilderKey{}).(*Builder)
	return builder
}

// WithBuilder returns a copy of parent ctx with builder associated.
// The associated builder can be accessed by WithCtx or Builder.Ctx.
func WithBuilder(ctx context.Context, builder *Builder) context.Context {
	return context.WithValue(ctx, ctxBuilderKey{}, builder)
}
