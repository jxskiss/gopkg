package zlog

import (
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ logr.LogSink = (*logrImpl)(nil)
var _ logr.CallDepthLogSink = (*logrImpl)(nil)

// NewLogrLogger creates a new logr.Logger.
func NewLogrLogger(options ...func(*LogrOptions)) logr.Logger {
	opts := newLogrOptions(options)
	l := opts.Logger.WithOptions(zap.AddCallerSkip(1))
	sink := &logrImpl{opts: opts, l: l}
	return logr.New(sink)
}

func newLogrOptions(options []func(*LogrOptions)) *LogrOptions {
	opts := &LogrOptions{
		DPanicOnInvalidLog: true,
	}
	for _, f := range options {
		f(opts)
	}
	// Set defaults.
	if opts.ErrorKey == "" {
		opts.ErrorKey = "error"
	}
	if opts.Logger == nil {
		opts.Logger = L().Logger
	}
	return opts
}

// LogrOptions customizes the behavior of logr logger created by NewLogrLogger.
type LogrOptions struct {
	// Logger optionally configures a zap.Logger to use instead of
	// the default logger.
	Logger *zap.Logger

	// ErrorKey replaces the default "error" field name used for the error
	// in Logger.Error calls.
	ErrorKey string

	// NumericLevelKey controls whether the numeric logr level is
	// added to each Info log message and the field key to use.
	NumericLevelKey string

	// DPanicOnInvalidLog controls whether extra log messages are emitted
	// for invalid log calls with zap's DPanic method.
	// Depending on the configuration of the zap logger, the program then
	// panics after emitting the log message which is useful in development
	// because such invalid log calls are bugs in the program.
	// The log messages explain why a call was invalid (for example,
	// non-string key, mismatched key-values pairs, etc.).
	// This is enabled by default.
	DPanicOnInvalidLog bool
}

type logrImpl struct {
	opts *LogrOptions
	l    *zap.Logger
}

func (r *logrImpl) Init(ri logr.RuntimeInfo) {
	r.l = r.l.WithOptions(zap.AddCallerSkip(ri.CallDepth))
}

const (
	// noLevel tells handleFields to not inject a numeric log level field.
	noLevel = -1
)

// handleFields is a slightly modified version of zap.SugaredLogger.sweetenFields.
func (r *logrImpl) handleFields(lv int, args []any, additional ...zap.Field) []zap.Field {
	injectNumLevel := r.opts.NumericLevelKey != "" && lv != noLevel

	if len(args) == 0 {
		// fast-return if we have no sugared fields and no "v" field
		if !injectNumLevel {
			return additional
		}
		return append(additional, zap.Int(r.opts.NumericLevelKey, lv))
	}

	// Unlike Zap, we can be pretty sure users aren't passing structured.
	// fields (since logr has no concept of that), so guess that we need a
	// little less space.
	numFields := len(args)/2 + len(additional)
	if injectNumLevel {
		numFields++
	}
	fields := make([]zap.Field, 0, numFields)
	if injectNumLevel {
		fields = append(fields, zap.Int(r.opts.NumericLevelKey, lv))
	}
	for i := 0; i < len(args); {
		// Check just in case for strongly-typed Zap fields,
		// which might be illegal (since it breaks implementation agnosticism).
		// If disabled, we can give a better error message.
		if f, ok := args[i].(zap.Field); ok {
			fields = append(fields, f)
			i++
			continue
		}

		// Make sure this isn't a mismatched key.
		if i == len(args)-1 {
			if r.opts.DPanicOnInvalidLog {
				r.l.WithOptions(zap.AddCallerSkip(1)).
					DPanic("odd number of arguments passed as key-value pairs for logging", toZapField("ignoredKey", args[i]))
			}
			break
		}

		// Process a key-value pair, ensuring that the key is a string.
		// If the key isn't a string, DPanic and stop checking the later arguments.
		key, val := args[i], args[i+1]
		rvKey := reflect.ValueOf(key)
		if rvKey.Kind() != reflect.String {
			if r.opts.DPanicOnInvalidLog {
				r.l.WithOptions(zap.AddCallerSkip(1)).
					DPanic("non-string key passed to logging, ignoring all later arguments", toZapField("invalidKey", key))
			}
			break
		}

		keyStr := rvKey.String()
		fields = append(fields, toZapField(keyStr, val))
		i += 2
	}

	return append(fields, additional...)
}

func toZapField(field string, val any) zap.Field {
	// Handle types that implement logr.Marshaler: log the replacement
	// object instead of the original one.
	if m, ok := val.(logr.Marshaler); ok {
		field, val = invokeLogrMarshaler(field, m)
	}
	return zap.Any(field, val)
}

func invokeLogrMarshaler(field string, m logr.Marshaler) (f string, ret any) {
	defer func() {
		if r := recover(); r != nil {
			ret = fmt.Sprintf("PANIC=%s", r)
			f = field + "Error"
		}
	}()
	return field, m.MarshalLog()
}

// Zap levels are int8, make sure we stay in bounds.
// logr itself should ensure we never get negative values.
func logrToZapLevel(lv int) zapcore.Level {
	if lv > 127 {
		lv = 127
	}
	return 0 - zapcore.Level(lv)
}

func (r *logrImpl) Enabled(lv int) bool {
	return r.l.Core().Enabled(logrToZapLevel(lv))
}

func (r *logrImpl) Info(lv int, msg string, keyAndVals ...any) {
	if ce := r.l.Check(logrToZapLevel(lv), msg); ce != nil {
		ce.Write(r.handleFields(lv, keyAndVals)...)
	}
}

func (r *logrImpl) Error(err error, msg string, keyAndVals ...any) {
	if ce := r.l.Check(zap.ErrorLevel, msg); ce != nil {
		ce.Write(r.handleFields(noLevel, keyAndVals, zap.NamedError(r.opts.ErrorKey, err))...)
	}
}

func (r *logrImpl) WithValues(keyAndValues ...any) logr.LogSink {
	clone := *r
	clone.l = r.l.With(r.handleFields(noLevel, keyAndValues)...)
	return &clone
}

func (r *logrImpl) WithName(name string) logr.LogSink {
	clone := *r
	clone.l = r.l.Named(name)
	return &clone
}

func (r *logrImpl) WithCallDepth(depth int) logr.LogSink {
	clone := *r
	clone.l = r.l.WithOptions(zap.AddCallerSkip(depth))
	return &clone
}

// Underlier exposes access to the underlying logging implementation.
// Since callers only have a logr.Logger, they have to know which
// implementation is in use, so this interface is less of an abstraction
// and more of way to test type conversion.
type Underlier interface {
	GetUnderlying() *zap.Logger
}

func (r *logrImpl) GetUnderlying() *zap.Logger {
	return r.l
}
