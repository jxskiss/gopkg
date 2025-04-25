package slog2zap

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ToZapLogger wraps a [slog.Logger] to a [zap.Logger].
func ToZapLogger(l *slog.Logger, opts ...zap.Option) *zap.Logger {
	core := ToZapCore(l)
	logger := zap.New(core, opts...)
	if h, ok := l.Handler().(withLoggerName); ok {
		logger = logger.Named(h.LoggerName())
	}
	return logger
}

// ToZapCore wraps a [slog.Logger] as a [zapcore.Core].
func ToZapCore(l *slog.Logger) zapcore.Core {
	core := &slogCore{handler: l.Handler()}
	if h, ok := l.Handler().(withLoggerName); ok {
		core.loggerName = h.LoggerName()
	}
	return core
}

type withLoggerName interface {
	LoggerName() string
}

// slogCore is a [zapcore.Core] implementation that forwards logs to [slog.Handler].
type slogCore struct {
	loggerName string
	handler    slog.Handler
}

func (c *slogCore) Enabled(level zapcore.Level) bool {
	return c.handler.Enabled(context.Background(), toSlogLevel(level))
}

func (c *slogCore) With(fields []zapcore.Field) zapcore.Core {
	handler := c.handler.WithAttrs(fieldToAttrs(fields))
	return &slogCore{handler: handler}
}

func (c *slogCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return ce.AddCore(entry, c)
	}
	return ce
}

func (c *slogCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	attrs := make([]slog.Attr, 0, len(fields)+3)
	if entry.LoggerName != "" {
		attrs = append(attrs, slog.String("logger", entry.LoggerName))
	} else if c.loggerName != "" {
		attrs = append(attrs, slog.String("logger", c.loggerName))
	}
	attrs = append(attrs, fieldToAttrs(fields)...)
	if entry.Stack != "" {
		attrs = append(attrs, slog.String("stacktrace", entry.Stack))
	}

	// https://pkg.go.dev/log/slog#hdr-Writing_a_handler
	r := slog.NewRecord(entry.Time, toSlogLevel(entry.Level), entry.Message, entry.Caller.PC)

	return c.handler.WithAttrs(attrs).Handle(context.Background(), r)
}

func (c *slogCore) Sync() error {
	return nil
}

func fieldToAttrs(fields []zapcore.Field) []slog.Attr {
	attrs := make([]slog.Attr, 0, len(fields))
	for _, field := range fields {
		attrs = append(attrs, fieldToAttr(field))
	}
	return attrs
}

func fieldToAttr(field zapcore.Field) slog.Attr {
	switch field.Type {
	case zapcore.StringType:
		return slog.String(field.Key, field.String)
	case zapcore.StringerType:
		return slog.String(field.Key, field.Interface.(fmt.Stringer).String())
	case zapcore.ByteStringType:
		return slog.String(field.Key, string(field.Interface.([]byte)))
	case zapcore.Int64Type, zapcore.Int32Type, zapcore.Int16Type, zapcore.Int8Type:
		return slog.Int64(field.Key, field.Integer)
	case zapcore.Uint64Type, zapcore.Uint32Type, zapcore.Uint16Type, zapcore.Uint8Type, zapcore.UintptrType:
		return slog.Uint64(field.Key, uint64(field.Integer))
	case zapcore.Float64Type:
		return slog.Float64(field.Key, math.Float64frombits(uint64(field.Integer)))
	case zapcore.Float32Type:
		return slog.Float64(field.Key, float64(math.Float32frombits(uint32(field.Integer))))
	case zapcore.BoolType:
		return slog.Bool(field.Key, field.Integer > 0)
	case zapcore.TimeType:
		if field.Interface != nil {
			loc, ok := field.Interface.(*time.Location)
			if ok {
				return slog.Time(field.Key, time.Unix(0, field.Integer).In(loc))
			}
		}
		return slog.Time(field.Key, time.Unix(0, field.Integer))
	case zapcore.TimeFullType:
		return slog.Time(field.Key, field.Interface.(time.Time))
	case zapcore.DurationType:
		return slog.Duration(field.Key, time.Duration(field.Integer))
	case zapcore.UnknownType, zapcore.NamespaceType, zapcore.SkipType:
		return slog.Attr{}
	default:
		return slog.Any(field.Key, field.Interface)
	}
}

// toSlogLevel converts a zapcore.Level to a slog.Level.
func toSlogLevel(level zapcore.Level) slog.Level {
	if level == zapcore.InfoLevel { // fast path
		return slog.LevelInfo
	}
	if level >= zapcore.ErrorLevel {
		return slog.LevelError
	}
	if level == zapcore.WarnLevel {
		return slog.LevelWarn
	}
	return slog.LevelDebug
}
