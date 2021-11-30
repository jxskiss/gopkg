package zlog

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

var disableTrace = false

// Trace logs a message at TraceLevel if it's enabled.
// It also adds a prefix "[TRACE] " to the message.
//
// If trace messages are disabled by GlobalConfig, calling this function
// is a no-op.
func Trace(msg string, fields ...zap.Field) {
	if !disableTrace {
		_l().Debug(TracePrefix+msg, fields...)
	}
}

// Tracef uses fmt.Sprintf to log a message at TraceLevel if it's enabled.
// It also adds a prefix "[TRACE] " to the message.
//
// If trace messages are disabled by GlobalConfig, calling this function
// is a no-op.
func Tracef(format string, args ...interface{}) {
	if !disableTrace {
		_s().Debugf(TracePrefix+format, args...)
	}
}

// TRACE logs a message at TraceLevel if it's enabled.
// It also adds a prefix "[TRACE] " to the message.
//
// TRACE accepts flexible arguments to help development, it trys to get a
// logger from the first argument, if the first argument is a *zap.Logger or
// *zap.SugaredLogger, the logger will be used, else if the first argument
// is a context.Context, the context will be used to build a logger using
// Builder, else it uses the global logger.
//
// The other arguments may be of type zap.Field or any ordinary type,
// the type will be detected and the arguments will be formatted in a most
// reasonable way. See example code for detailed usage examples.
//
// If trace messages are disabled by GlobalConfig, calling this function
// is a no-op.
func TRACE(args ...interface{}) {
	if !disableTrace {
		_slowPathTRACE(0, args...)
	}
}

// TRACESkip is similar to TRACE, but it has an extra skip argument to get
// correct caller information. When you need to wrap TRACE, you will always
// want to use this function instead of TRACE.
//
// If trace messages are disabled by GlobalConfig, calling this function
// is a no-op.
func TRACESkip(skip int, args ...interface{}) {
	if !disableTrace {
		_slowPathTRACE(skip, args...)
	}
}

func _slowPathTRACE(skip int, args ...interface{}) {
	logger, msg, fields := parseLoggerAndParams(skip, args)
	msg = addCallerPrefix(skip, TracePrefix, msg)
	logger.Debug(msg, fields...)
}

func parseLoggerAndParams(skip int, args []interface{}) (*zap.Logger, string, []zap.Field) {
	var logger = L()
	if len(args) > 0 {
		switch arg0 := args[0].(type) {
		case context.Context:
			logger = B(arg0).Build()
			args = args[1:]
		case *zap.Logger:
			logger = arg0
			args = args[1:]
		case *zap.SugaredLogger:
			logger = arg0.Desugar()
			args = args[1:]
		}
	}
	logger = logger.WithOptions(zap.AddCallerSkip(skip + 2))
	if len(args) == 0 {
		return logger, "", nil
	}

	switch arg0 := args[0].(type) {
	case string:
		fields, ok := tryConvertFields(args[1:])
		if ok {
			return logger, arg0, fields
		}
	case zap.Field:
		fields, ok := tryConvertFields(args)
		if ok {
			return logger, "", fields
		}
	}

	template := ""
	if s, ok := args[0].(string); ok && strings.IndexByte(s, '%') >= 0 {
		template = s
		args = args[1:]
		if len(args) == 0 {
			return logger, template, nil
		}
	}
	return logger, formatMessage(template, args), nil
}

func addCallerPrefix(skip int, prefix, msg string) string {
	caller, file, line, _ := getCaller(skip + 3)
	if msg == "" {
		return fmt.Sprintf("%s========  %s#L%d - %s  ========", prefix, file, line, caller)
	}
	return fmt.Sprintf("%s[%s] %s", prefix, caller, msg)
}

func tryConvertFields(args []interface{}) ([]zap.Field, bool) {
	if len(args) == 0 {
		return nil, true
	}
	for i := 0; i < len(args); i++ {
		if _, ok := args[i].(zap.Field); !ok {
			return nil, false
		}
	}
	fields := make([]zap.Field, len(args))
	for i, f := range args {
		fields[i] = f.(zap.Field)
	}
	return fields, true
}

func formatMessage(template string, fmtArgs []interface{}) string {
	if len(fmtArgs) == 0 {
		return template
	}
	if template != "" {
		return fmt.Sprintf(template, fmtArgs...)
	}
	if len(fmtArgs) == 1 {
		if str, ok := fmtArgs[0].(string); ok {
			return str
		}
	}
	template = "%v" + strings.Repeat(" %v", len(fmtArgs)-1)
	return fmt.Sprintf(template, fmtArgs...)
}
