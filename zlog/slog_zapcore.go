//go:build go1.21

package zlog

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"go.uber.org/zap/zapcore"
)

// slogCoreImpl is a zapcore.Core implementation that forwards logs to
// slog.Handler.
type slogCoreImpl struct {
	cfg     *Config
	handler slog.Handler
}

func (c *slogCoreImpl) Enabled(level zapcore.Level) bool {
	return c.handler.Enabled(context.Background(), zapToSlogLevel(level))
}

func fieldToAttr(field zapcore.Field) slog.Attr {
	switch field.Type {
	case zapcore.StringType:
		return slog.String(field.Key, field.String)
	case zapcore.Int64Type:
		return slog.Int64(field.Key, field.Integer)
	case zapcore.Int32Type:
		return slog.Int(field.Key, int(field.Integer))
	case zapcore.Uint64Type:
		return slog.Uint64(field.Key, uint64(field.Integer))
	case zapcore.Float64Type:
		return slog.Float64(field.Key, math.Float64frombits(uint64(field.Integer)))
	case zapcore.BoolType:
		return slog.Bool(field.Key, field.Integer == 1)
	case zapcore.TimeType:
		if field.Interface != nil {
			loc, ok := field.Interface.(*time.Location)
			if ok {
				return slog.Time(field.Key, time.Unix(0, field.Integer).In(loc))
			}
		}
		return slog.Time(field.Key, time.Unix(0, field.Integer))
	case zapcore.DurationType:
		return slog.Duration(field.Key, time.Duration(field.Integer))
	default:
		return slog.Any(field.Key, field.Interface)
	}
}

func fieldToAttrs(fields []zapcore.Field) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(fields))
	for _, field := range fields {
		attrs = append(attrs, fieldToAttr(field))
	}
	return attrs
}

func (c *slogCoreImpl) With(fields []zapcore.Field) zapcore.Core {
	return &slogCoreImpl{handler: c.handler.WithAttrs(fieldToAttrs(fields))}
}

func (c *slogCoreImpl) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return ce.AddCore(entry, c)
	}
	return ce
}

func (c *slogCoreImpl) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	attrs := make([]slog.Attr, 0, len(fields)+3)
	if entry.LoggerName != "" {
		attrs = append(attrs, slog.String("logger", entry.LoggerName))
	}
	if !c.cfg.DisableCaller && entry.Caller.Defined {
		attrs = append(attrs, slog.Any("source", entry.Caller))
	}
	if !c.cfg.DisableStacktrace && entry.Stack != "" {
		attrs = append(attrs, slog.String("stacktrace", entry.Stack))
	}
	attrs = append(attrs, fieldToAttrs(fields)...)

	// https://pkg.go.dev/log/slog#hdr-Writing_a_handler
	r := slog.NewRecord(entry.Time, zapToSlogLevel(entry.Level), entry.Message, 0)
	r.AddAttrs(attrs...)

	err := c.handler.Handle(context.Background(), r)
	if err != nil {
		return fmt.Errorf("write log: %w", err)
	}

	return nil
}

func (c *slogCoreImpl) Sync() error {
	return nil
}

// zapToSlogLevel converts a zapcore.Level to a slog.Level.
// unsupported levels are converted to slog.LevelDebug.
func zapToSlogLevel(level zapcore.Level) slog.Level {
	switch level {
	case zapcore.DebugLevel:
		return slog.LevelDebug
	case zapcore.InfoLevel:
		return slog.LevelInfo
	case zapcore.WarnLevel:
		return slog.LevelWarn
	case zapcore.ErrorLevel:
		return slog.LevelError
	case zapcore.DPanicLevel:
		return slog.LevelError
	case zapcore.PanicLevel:
		return slog.LevelError
	case zapcore.FatalLevel:
		return slog.LevelError
	default:
		return slog.LevelDebug
	}
}
