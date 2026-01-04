package zlog

import (
	"context"
	"log/slog"

	"github.com/go-logr/logr"
)

// NewCtx returns a copy of ctx with the logger attached.
// The parent context will be unaffected.
func NewCtx(parent context.Context, logger *Logger) context.Context {
	if parent == nil {
		parent = context.Background()
	}
	return logr.NewContextWithSlogLogger(parent, logger)
}

// FromCtx returns the logger associated with the ctx.
// If no logger is associated, or the logger or ctx is nil,
// slog.Default() is used.
// The returned logger's handler is a *Handler with fromCtx set to ctx,
// which is used when Handler.Handle is called without ctx.
// This function will convert a logr.Logger to a *slog.Logger only if necessary.
func FromCtx(ctx context.Context) *Logger {
	if ctx == nil {
		ctx = context.Background()
	}
	return slog.New(fromCtxHandler(ctx))
}

// fromCtxHandler helps to optimize performance of scoped loggers.
func fromCtxHandler(ctx context.Context) *Handler {
	h0 := fromCtx(ctx).Handler()
	if h1, ok := h0.(*Handler); ok {
		return h1.withContext(ctx)
	}
	return &Handler{
		next:    h0,
		fromCtx: ctx,
	}
}

func fromCtx(ctx context.Context) *Logger {
	l := logr.FromContextAsSlogLogger(ctx)
	if l != nil {
		return l
	}
	return slog.Default()
}
