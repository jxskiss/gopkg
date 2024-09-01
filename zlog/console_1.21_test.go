//go:build go1.21

package zlog

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestCoreForConsole(t *testing.T) {
	cfg := checkAndFillDefaults(&Config{Development: true})
	enc, err := cfg.buildEncoder()
	require.NoError(t, err)

	buf := zaptest.Buffer{}
	core := newCoreForConsole(&Config{Development: true}, enc, &buf)

	logger, _, err := NewWithCore(&WrapCoreConfig{Level: DebugLevel}, core)
	require.NoError(t, err)

	slogger := NewSlogLogger(func(options *SlogOptions) {
		options.Logger = logger
	})

	logger.Debug("zap logger debug message", zap.String("key1", "value1"), zap.Error(errors.New("test error")))
	slogger.Debug("slog logger debug message", "key1", "value1", "error", errors.New("test error"))

	logger.Info("zap logger info message", zap.String("key1", "value1"), zap.Error(errors.New("test error")))
	slogger.Info("slog logger info message", "key1", "value1", "error", errors.New("test error"))

	got := buf.String()
	assert.Contains(t, got, `DEBUG  zap logger debug message 	error= "test error"  key1= value1`)
	assert.Contains(t, got, `DEBUG  slog logger debug message 	error= "test error"  key1= value1`)
	assert.Contains(t, got, `INFO   zap logger info message 	error= "test error"  key1= value1`)
	assert.Contains(t, got, `INFO   slog logger info message 	error= "test error"  key1= value1`)
}
