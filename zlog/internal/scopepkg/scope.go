package scopepkg

import (
	"context"
	"log/slog"
)

var scopeLog iScope

type iScope interface {
	Logger() *slog.Logger
	With(ctx context.Context, args ...any) *slog.Logger
	WithError(ctx context.Context, err error, args ...any) *slog.Logger
	WithGroup(ctx context.Context, group string, args ...any) *slog.Logger
}

func SetupScoperLogger(s iScope) {
	scopeLog = s
	_ = scopeLog
}

func PrintLog(ctx context.Context, args ...any) {
	scopeLog.With(ctx).InfoContext(ctx, "test scope logger", args...)
}
