//go:build go1.21

package zlog

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestSlogCoreImpl(t *testing.T) {
	t.Parallel()

	var b bytes.Buffer
	handler := slog.NewTextHandler(
		io.MultiWriter(&b, testLogWriter{t}),
		&slog.HandlerOptions{Level: slog.LevelDebug},
	)

	loggerZap, err := zap.NewProduction(zap.WrapCore(func(zapcore.Core) zapcore.Core {
		return &slogCoreImpl{
			cfg:     &Config{Development: true},
			handler: handler,
		}
	}))
	require.NoError(t, err)

	loggerZap.Debug("debug level")
	loggerZap.Info("info level")
	loggerZap.Warn("warn level")
	loggerZap.Error("error level")

	err = loggerZap.Sync()
	require.NoError(t, err)

	bs := b.String()

	assert.Contains(t, bs, `level=DEBUG msg="debug level"`)
	assert.Contains(t, bs, `level=INFO msg="info level"`)
	assert.Contains(t, bs, `level=WARN msg="warn level"`)
	assert.Contains(t, bs, `level=ERROR msg="error level"`)
}

func BenchmarkSlog_SlogCoreImpl(b *testing.B) {
	handler := noopSlogHandler{}
	loggerZap, err := zap.NewProduction(zap.WrapCore(func(zapcore.Core) zapcore.Core {
		return &slogCoreImpl{
			cfg:     &Config{Development: true},
			handler: handler,
		}
	}))
	require.NoError(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		loggerZap.Info("hello world")
	}
}

type noopSlogHandler struct{}

func (noopSlogHandler) Enabled(context.Context, slog.Level) bool  { return true }
func (noopSlogHandler) Handle(context.Context, slog.Record) error { return nil }
func (h noopSlogHandler) WithAttrs([]slog.Attr) slog.Handler      { return h }
func (h noopSlogHandler) WithGroup(string) slog.Handler           { return h }

type testLogWriter struct{ t *testing.T }

func (w testLogWriter) Write(p []byte) (int, error) {
	w.t.Log(strings.TrimSuffix(string(p), "\n"))

	return len(p), nil
}
