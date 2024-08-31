//go:build !go1.21

package zlog

import "go.uber.org/zap/zapcore"

func newCoreForConsole(_ *Config, enc zapcore.Encoder, ws zapcore.WriteSyncer) zapcore.Core {
	return zapcore.NewCore(enc, ws, Level(-127))
}
