package zlog

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// B returns an empty Builder.
func B() *Builder { return &Builder{} }

// Builder provides chaining methods to build a logger.
// Different with calling zap.Logger's methods, it does not write the context
// information to underlying buffer immediately, later data will override
// the former which has key. When the builder is prepared, call Build to get
// the final logger.
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
// the old ones.
func (b *Builder) With(fields ...zap.Field) *Builder {
	if len(fields) == 0 {
		return b
	}
	out := b.clone()
	out.fields = make([]zap.Field, len(b.fields), len(b.fields)+len(fields))
	copy(out.fields, b.fields)
loop:
	for _, f := range fields {
		for i, x := range out.fields {
			if f.Key == x.Key {
				out.fields[i] = f
				continue loop
			}
		}
		out.fields = append(out.fields, f)
	}
	return out
}

// Ctx adds fields extracted from ctx to the builder.
//
// Note: to use this, Config.CtxFunc must be set, else it logs and error
// message at DPANIC level.
func (b *Builder) Ctx(ctx context.Context) *Builder {
	ctxFunc := gP.cfg.CtxFunc
	if ctxFunc == nil {
		L().DPanic("calling Builder.Ctx without CtxFunc configured")
		return b
	}
	if ctx == nil {
		return nil
	}
	return b.With(ctxFunc(ctx)...)
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
