//go:build go1.21

package zlog

import (
	"io"
	"log/slog"
	"os"
	"time"

	slogconsolehandler "github.com/jxskiss/slog-console-handler"
	"go.uber.org/zap/zapcore"
)

func newCoreForConsole(cfg *Config, _ zapcore.Encoder, ws zapcore.WriteSyncer) zapcore.Core {
	opts := &slogconsolehandler.HandlerOptions{
		// NB: caller info is handled by the ReplaceAttr option.
		AddSource: false,
		Level:     slog.Level(-127),
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
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
					if s, ok := a.Value.Any().(*slog.Source); ok && s != nil {
						if s.File == "" {
							return slog.Attr{}
						}
						return slog.String(slog.SourceKey, slogconsolehandler.FormatSourceShort(*s))
					}
				}
			}
			return a
		},
		DisableColor: false,
	}
	writer := io.Writer(ws)
	if _, ok := ws.(*wrapStderr); ok {
		writer = os.Stderr
	}
	impl := &slogCoreImpl{
		cfg:     cfg,
		handler: slogconsolehandler.New(writer, opts),
	}
	return impl
}
