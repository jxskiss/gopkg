//go:build go1.21

package zlog

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCoreForConsole(t *testing.T) {
	cfg := checkAndFillDefaults(&Config{Development: true})
	enc, err := cfg.buildEncoder()
	require.NoError(t, err)
	output, _, err := zap.Open("stderr")
	require.NoError(t, err)
	core := newCoreForConsole(&Config{Development: true}, enc, output)

	logger, _, err := NewWithCore(&WrapCoreConfig{Level: DebugLevel}, core)
	require.NoError(t, err)

	slogger := NewSlogLogger(func(options *SlogOptions) {
		options.Logger = logger
	})

	logger.Debug("zap logger debug message", zap.String("key1", "value1"), zap.Error(errors.New("test error")))
	slogger.Debug("slog logger debug message", "key1", "value1", "error", errors.New("test error"))

	logger.Info("zap logger info message", zap.String("key1", "value1"), zap.Error(errors.New("test error")))
	slogger.Info("slog logger info message", "key1", "value1", "error", errors.New("test error"))
}
