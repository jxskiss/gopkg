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
	l := fromCtx(ctx)
	if h, ok := l.Handler().(*Handler); ok {
		if h.fromCtx == ctx {
			return l
		}
		return slog.New(h.withContext(ctx))
	}
	h := &Handler{
		next:    l.Handler(),
		fromCtx: ctx,
	}
	return slog.New(h)
}

func fromCtx(ctx context.Context) *Logger {
	l := logr.FromContextAsSlogLogger(ctx)
	if l != nil {
		return l
	}
	return slog.Default()
}
