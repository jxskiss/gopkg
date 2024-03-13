//go:build go1.21

package zlog

import (
	"context"
	"errors"
	"log"
	"log/slog"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func getErrorWithStackTrace() error {
	return &errorWithStack{
		error: errors.New("test error"),
		stack: debug.Stack(),
	}
}

type errorWithStack struct {
	error
	stack []byte
}

func TestRedirectStdLog(t *testing.T) {
	out := &zaptest.Buffer{}
	l, _, err := NewWithOutput(&Config{Level: "debug", Format: "console"}, out)
	require.Nil(t, err)
	resetFunc := redirectStdLog(l, false)
	defer resetFunc()

	log.Printf("[Debug] std log Printf, key1=%v", "value1")
	slog.Debug("std slog Debug", "key1", "value1")

	lines := out.Lines()
	require.Len(t, lines, 2)
	assert.Contains(t, lines[0], "DEBUG stdlog zlog/slog_test.go:")
	assert.Contains(t, lines[0], "[Debug] std log Printf, key1=value1")
	assert.Contains(t, lines[1], "DEBUG zlog/slog_test.go:")
	assert.Contains(t, lines[1], `std slog Debug {"key1": "value1"}`)
}

func TestSetSlogDefault(t *testing.T) {
	out := &zaptest.Buffer{}
	l, p, err := NewWithOutput(&Config{Level: "warn", Format: "console"}, out)
	require.Nil(t, err)
	resetFunc := ReplaceGlobals(l, p)
	defer resetFunc()

	SetSlogDefault(NewSlogLogger())
	log.Printf("[Warn] std log Printf, key1=%v", "value1")
	slog.Info("std slog Info", "key1", "value1")
	slog.Warn("std slog Warn", "key2", "value2")

	lines := out.Lines()
	require.Len(t, lines, 2)
	assert.Contains(t, lines[0], "WARN stdlog zlog/slog_test.go:")
	assert.Contains(t, lines[0], "[Warn] std log Printf, key1=value1")
	assert.Contains(t, lines[1], "WARN zlog/slog_test.go:")
	assert.Contains(t, lines[1], `std slog Warn {"key2": "value2"}`)
}

func TestNewSlogLogger(t *testing.T) {
	r0 := NewSlogLogger()
	assert.NotNil(t, r0.Handler().(*slogImpl).opts)
	assert.NotNil(t, r0.Handler().(*slogImpl).l)

	l := L().Named("slogTest").With(zap.String("ns", "default"))
	r1 := NewSlogLogger(func(options *SlogOptions) {
		options.Logger = l.Logger
	})
	assert.Equal(t, "slogTest", r1.Handler().(*slogImpl).name)
	r1.Info("test NewSlogLogger with logger")
}

func TestSlogLoggerLogging(t *testing.T) {
	out := &zaptest.Buffer{}
	l, _, err := NewWithOutput(&Config{
		Level:             "trace",
		Format:            "json",
		DisableTimestamp:  true,
		DisableStacktrace: true,
	}, out)
	require.Nil(t, err)
	logger := NewSlogLogger(func(options *SlogOptions) {
		options.Logger = l
	})

	logger.Info("test info", "ns", "default", "podnum", 2)
	lines := out.Lines()
	assert.Contains(t, lines[0], `"level":"info"`)
	assert.Contains(t, lines[0], `"caller":"zlog/slog_test.go:`)
	assert.Contains(t, lines[0], `"ns":"default"`)
	assert.Contains(t, lines[0], `"podnum":2`)
}

func TestSlogLoggerCtxHandler(t *testing.T) {
	t.Run("change level", func(t *testing.T) {
		out := &zaptest.Buffer{}
		cfg := &Config{Level: "info", Format: "console"}
		cfg.RedirectStdLog = true
		cfg.CtxHandler.ChangeLevel = func(ctx context.Context) *Level {
			if ctx.Value("debug") != nil {
				level := DebugLevel
				return &level
			}
			return nil
		}
		cfg.CtxHandler.WithCtx = func(ctx context.Context) (result CtxResult) {
			if ctx.Value("debug") != nil {
				level := DebugLevel
				result.Level = &level
			}
			return
		}
		l, p, err := NewWithOutput(cfg, out)
		require.Nil(t, err)
		resetFunc := ReplaceGlobals(l, p)
		defer resetFunc()

		ctx := context.WithValue(context.Background(), "debug", "1")
		slog.DebugContext(ctx, "a debug message")

		lines := out.Lines()
		require.Len(t, lines, 1)
		assert.Contains(t, lines[0], "DEBUG zlog/slog_test.go:")
		assert.Contains(t, lines[0], `a debug message`)
	})

	t.Run("ctx fields", func(t *testing.T) {
		out := &zaptest.Buffer{}
		cfg := &Config{Level: "info", Format: "console"}
		cfg.RedirectStdLog = true
		cfg.CtxHandler.WithCtx = func(ctx context.Context) (result CtxResult) {
			result.Fields = append(result.Fields,
				zap.String("logid", "abcde"),
				zap.Int64("userID", 12345))
			return
		}
		l, p, err := NewWithOutput(cfg, out)
		require.Nil(t, err)
		resetFunc := ReplaceGlobals(l, p)
		defer resetFunc()

		slog.InfoContext(context.Background(), "an info message")

		lines := out.Lines()
		require.Len(t, lines, 1)
		assert.Contains(t, lines[0], "INFO zlog/slog_test.go:")
		assert.Contains(t, lines[0], `an info message {"logid": "abcde", "userID": 12345}`)
	})
}

func TestSlogOptionsReplaceAttr(t *testing.T) {
	t.Run("error key", func(t *testing.T) {
		out := &zaptest.Buffer{}
		l, _, err := NewWithOutput(&Config{Level: "debug", Format: "console"}, out)
		require.Nil(t, err)
		logger := NewSlogLogger(func(options *SlogOptions) {
			options.Logger = l
			options.ReplaceAttr = func(groups []string, a slog.Attr) (rr ReplaceResult) {
				if a.Key == "error" || a.Key == "err" && a.Value.Kind() == slog.KindAny {
					if err, ok := a.Value.Any().(error); ok {
						rr.Field = zap.Error(err)
						return
					}
				}
				rr.Field = ConvertAttrToField(a)
				return
			}
		})

		logger.Error("error log 1", "err", errors.New("test error 1"))
		logger.Error("error log 2", "error", errors.New("test error 2"))
		logger.Error("error log 3", "notMatchErrKey", errors.New("test error 3"))

		lines := out.Lines()
		require.Len(t, lines, 3)
		assert.Contains(t, lines[0], "ERROR zlog/slog_test.go:")
		assert.Contains(t, lines[0], `error log 1 {"error": "test error 1"}`)
		assert.Contains(t, lines[1], "ERROR zlog/slog_test.go:")
		assert.Contains(t, lines[1], `error log 2 {"error": "test error 2"}`)
		assert.Contains(t, lines[2], "ERROR zlog/slog_test.go:")
		assert.Contains(t, lines[2], `error log 3 {"notMatchErrKey": "test error 3"}`)
	})

	t.Run("error stacktrace", func(t *testing.T) {
		out := &zaptest.Buffer{}
		l, _, err := NewWithOutput(&Config{Level: "debug", Format: "console"}, out)
		require.Nil(t, err)
		logger := NewSlogLogger(func(options *SlogOptions) {
			options.Logger = l
			options.ReplaceAttr = func(groups []string, a slog.Attr) (rr ReplaceResult) {
				if a.Key == "error" || a.Key == "err" && a.Value.Kind() == slog.KindAny {
					if err, ok := a.Value.Any().(error); ok {
						if inner, ok := err.(*errorWithStack); ok {
							rr.Multi = []zap.Field{
								zap.Error(err),
								zap.ByteString("stacktrace", inner.stack),
							}
						} else {
							rr.Field = zap.Error(err)
						}
						return
					}
				}
				rr.Field = ConvertAttrToField(a)
				return
			}
		})

		logger.Error("error log", "err", wrapFunc(3, getErrorWithStackTrace)())

		lines := out.Lines()
		require.Len(t, lines, 1)
		assert.Contains(t, lines[0], "ERROR zlog/slog_test.go:")
		assert.Contains(t, lines[0], `error log {"error": "test error", "stacktrace": "`)
		assert.Contains(t, lines[0], `zlog.getErrorWithStackTrace()\n\t`)
		assert.Contains(t, lines[0], "zlog/slog_test.go:23")
	})
}

func wrapFunc(depth int, f func() error) func() error {
	for i := 0; i < depth; i++ {
		copyFunc := f
		newFunc := func() error {
			return copyFunc()
		}
		f = newFunc
	}
	return f
}
