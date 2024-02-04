package zlog

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

func TestNew(t *testing.T) {
	logger, props, err := New(&Config{Level: "trace", Format: "logfmt"})
	require.Nil(t, err)
	require.NotNil(t, logger)
	require.NotNil(t, props)
	defer props.CloseWriters()

	Logger{Logger: logger}.Trace("trace message")
	logger.Info("info message")
}

func TestNewWithCore(t *testing.T) {
	cfg := &WrapCoreConfig{
		Level: TraceLevel,
	}
	core := zapcore.NewNopCore()
	logger, props, err := NewWithCore(cfg, core)
	require.Nil(t, err)
	require.NotNil(t, logger)
	require.NotNil(t, props)
	defer props.CloseWriters()

	Logger{Logger: logger}.Trace("trace message")
	logger.Info("info message")
}
