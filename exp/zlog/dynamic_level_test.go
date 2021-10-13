package zlog

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestDynamicLevelCore(t *testing.T) {
	buf := &bytes.Buffer{}
	logger, _, err := NewWithOutput(&Config{Development: false, Level: "error"}, zapcore.AddSync(buf))
	assert.Nil(t, err)

	// assert we get a dynamicLevelCore
	_ = logger.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		_, ok := core.(*dynamicLevelCore)
		assert.True(t, ok)
		return core
	}))

	// the level is "error", these messages won't be logged
	logger.Debug("debug message 1")
	logger.Info("info message 1")
	logger.Warn("warn message 1")

	// this error message will be logged
	logger.Error("error message 1")

	got1 := buf.String()
	assert.NotContains(t, got1, "debug message 1")
	assert.NotContains(t, got1, "info message 1")
	assert.NotContains(t, got1, "warn message 1")
	assert.Contains(t, got1, "error message 1")

	// change level to debug
	logger = B(nil).Base(logger).Level(DebugLevel).Build()
	logger.Debug("debug message 2")
	logger.Info("info message 2")
	logger.Warn("warn message 2")
	logger.Error("error message 2")

	got2 := buf.String()
	assert.Contains(t, got2, "debug message 2")
	assert.Contains(t, got2, "info message 2")
	assert.Contains(t, got2, "warn message 2")
	assert.Contains(t, got2, "error message 2")
}

func TestDynamicLevelCore_ChangeLevelWithCtx(t *testing.T) {
	var buf = &bytes.Buffer{}
	var _replace = func(buf *bytes.Buffer) func() {
		cfg := &Config{
			Level: "warn",
			CtxFunc: func(ctx context.Context, args CtxArgs) (result CtxResult) {
				if ctx.Value("level") != nil {
					level := ctx.Value("level").(Level)
					result.Level = &level
				}
				return result
			},
		}
		l, p, err := NewWithOutput(cfg, zapcore.AddSync(buf))
		if err != nil {
			panic(err)
		}
		return replaceGlobals(l, p)
	}
	defer _replace(buf)()

	// the level is "warn", this info message won't be logged
	ctx1 := context.Background()
	WithCtx(ctx1).Info("info message 1")

	got1 := buf.String()
	assert.NotContains(t, got1, "info message 1")

	// set level to "info" from ctx, info messages will be logged
	ctx2 := context.WithValue(context.Background(), "level", InfoLevel)
	WithCtx(ctx2).Debug("debug message 2")
	WithCtx(ctx2).Info("info message 2")

	got2 := buf.String()
	assert.NotContains(t, got2, "debug message 2")
	assert.Contains(t, got2, "info message 2")
}

var (
	benchmarkMessage = "some test debug message is not too long and not too short"
	benchmarkFields  = []zap.Field{
		zap.String("key1", "value1"),
		zap.String("some_key_2", "some_value2"),
		zap.Int64("some_key_3", 183491839141471),
		zap.Int64s("some_slice_key_4", []int64{18412, 1312490194301, 318431849, 18912438918941}),
	}
)

func BenchmarkZapIoCore(b *testing.B) {
	logger := newBenchmarkZapIoCoreLogger()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Debug(benchmarkMessage, benchmarkFields...)
	}
}

func BenchmarkDynamicLevelCoreOverhead(b *testing.B) {
	logger := newBenchmarkDynamicLevelCoreLogger()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		logger.Debug(benchmarkMessage, benchmarkFields...)
	}
}

type discardWriter struct{}

func (w *discardWriter) Write(p []byte) (n int, err error) {
	return len(p), nil
}

func (w *discardWriter) Sync() error {
	return nil
}

func newBenchmarkZapIoCoreLogger() *zap.Logger {
	logger, _, err := NewWithOutput(&Config{Development: false, Level: "debug"}, &discardWriter{})
	if err != nil {
		panic(err)
	}

	// Make sure we unwrap the dynamic level wrapper.
	logger = logger.WithOptions(zap.WrapCore(unwrapDynamicLevelCore))
	_ = logger.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		if _, ok := core.(*dynamicLevelCore); ok {
			panic("core is wrapped by dynamicLevelCore")
		}
		return core
	}))
	return logger
}

func newBenchmarkDynamicLevelCoreLogger() *zap.Logger {
	logger, _, err := NewWithOutput(&Config{Development: false, Level: "warn"}, &discardWriter{})
	if err != nil {
		panic(err)
	}

	// Make sure we get a wrapped dynamic level core.
	logger = B(nil).Base(logger).Level(DebugLevel).Build()
	if err != nil {
		panic(err)
	}
	_ = logger.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		if _, ok := core.(*dynamicLevelCore); !ok {
			panic("core is not wrapped by dynamicLevelCore")
		}
		return core
	}))
	return logger
}
