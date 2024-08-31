//go:build go1.21

package zlog

import (
	"log/slog"
	"os"
	"time"

	"github.com/jxskiss/slog-console-handler"
	"go.uber.org/zap/zapcore"
)

func newCoreForConsole(cfg *Config, enc zapcore.Encoder, ws zapcore.WriteSyncer) zapcore.Core {
	opts := &slogconsolehandler.HandlerOptions{
		// NB: caller info is handled by the ReplaceAttr option.
		AddSource: false,
		Level:     slog.Level(-127),
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				if a.Value.Kind() == slog.KindTime {
					if t, ok := a.Value.Any().(time.Time); ok {
						if t.IsZero() {
							return slog.Attr{}
						}
						return slog.String(slog.TimeKey, slogconsolehandler.FormatTimeShort(t))
					}
				}
			case slog.SourceKey:
				if a.Value.Kind() == slog.KindAny {
					if s, ok := a.Value.Any().(zapcore.EntryCaller); ok {
						if !s.Defined {
							return slog.Attr{}
						}
						return slog.String(slog.SourceKey, slogconsolehandler.FormatSourceShort(slog.Source{
							Function: s.Function,
							File:     s.File,
							Line:     s.Line,
						}))
					}
				}
			}
			return a
		},
		DisableColor: false,
	}
	impl := &slogCoreImpl{
		cfg:     cfg,
		handler: slogconsolehandler.New(os.Stderr, opts),
	}
	return impl
}