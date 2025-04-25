package slog2zap

import (
	"bytes"
	"fmt"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestWrappedLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	slogLogger := slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{AddSource: true, Level: slog.LevelDebug}))

	zapLogger := ToZapLogger(slogLogger, zap.AddCaller())
	zapLogger.Debug("debug message 1", zap.String("k1", "value1"))
	zapLogger.Info("info message 1", zap.Int("k1", 54321))
	zapLogger.Info("info message 2", zap.Stringer("k1", &testStruct{val: "stringer testStruct val"}))
	zapLogger.Warn("warn message 1", zap.Time("time1", time.Now()))
	zapLogger.Error("error message 1", zap.Duration("cost", time.Second))

	got := buf.String()
	assert.Regexp(t, `time=\d{4}-\d{2}-\d{2}.{8,} level=DEBUG source=.*/zlog/slog2zap/wrapper_test.go:\d+ msg="debug message 1" k1=value1`, got)
	assert.Regexp(t, `time=\d{4}-\d{2}-\d{2}.{8,} level=INFO source=.*/zlog/slog2zap/wrapper_test.go:\d+ msg="info message 1" k1=54321`, got)
	assert.Regexp(t, `time=\d{4}-\d{2}-\d{2}.{8,} level=INFO source=.*/zlog/slog2zap/wrapper_test.go:\d+ msg="info message 2" k1="stringer testStruct val"`, got)
	assert.Regexp(t, `time=\d{4}-\d{2}-\d{2}.{8,} level=WARN source=.*/zlog/slog2zap/wrapper_test.go:\d+ msg="warn message 1" time1=\d{4}-\d{2}-\d{2}.{8,}`, got)
	assert.Regexp(t, `time=\d{4}-\d{2}-\d{2}.{8,} level=ERROR source=.*/zlog/slog2zap/wrapper_test.go:\d+ msg="error message 1" cost=1s`, got)
}

func TestWithLoggerName(t *testing.T) {
	buf := &bytes.Buffer{}
	slogHandler := &namedHandler{
		loggerName: "testNamedLogger",
		Handler:    slog.NewTextHandler(buf, nil),
	}
	slogLogger := slog.New(slogHandler)

	zapLogger := ToZapLogger(slogLogger)
	zapLogger.Info("test log message")

	got := buf.String()
	assert.Regexp(t, `time=\d{4}-\d{2}-\d{2}.{8,} level=INFO msg="test log message" logger=testNamedLogger`, got)
}

type testStruct struct {
	val any
}

func (x *testStruct) String() string {
	return fmt.Sprint(x.val)
}

type namedHandler struct {
	loggerName string
	slog.Handler
}

func (h *namedHandler) LoggerName() string {
	return h.loggerName
}
