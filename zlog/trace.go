package zlog

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"

	"github.com/jxskiss/gopkg/v2/internal/logfilter"
)

const TraceFilterRuleEnvName = "ZLOG_TRACE_FILTER_RULE"

func (p *Properties) compileTraceFilter() {
	if p.cfg.TraceFilterRule == "" {
		envRule := os.Getenv(TraceFilterRuleEnvName)
		if envRule != "" {
			S().Infof("zlog: using trace filter rule from env: %q", envRule)
			p.cfg.TraceFilterRule = envRule
		}
	}
	if p.cfg.TraceFilterRule != "" {
		var errs []error
		p.traceFilter, errs = logfilter.NewFileNameFilter(p.cfg.TraceFilterRule)
		for _, err := range errs {
			S().Warnf("zlog: %v", err)
		}
	}
}

// Trace logs a message at TraceLevel if it's enabled.
// It also adds a prefix "[TRACE] " to the message.
//
// If trace messages are disabled globally, calling this function is
// a no-op.
func (l Logger) Trace(msg string, fields ...zap.Field) {
	if globals.Level.Load() <= int32(TraceLevel) {
		l.slowPathTrace(msg, fields)
	}
}

func (l Logger) slowPathTrace(msg string, fields []zap.Field) {
	logger := l.Logger.WithOptions(zap.AddCallerSkip(3))
	checkAndWriteTraceMessage(logger, msg, fields...)
}

// Tracef uses fmt.Sprintf to log a message at TraceLevel if it's enabled.
// It also adds a prefix "[TRACE] " to the message.
//
// If trace messages are disabled globally, calling this function is
// a no-op.
func (l Logger) Tracef(format string, args ...any) {
	if globals.Level.Load() <= int32(TraceLevel) {
		l.slowPathTracef(format, args)
	}
}

func (l Logger) slowPathTracef(format string, args []any) {
	logger := l.Logger.WithOptions(zap.AddCallerSkip(3))
	msg := formatMessage(format, args)
	checkAndWriteTraceMessage(logger, msg)
}

// Tracef uses fmt.Sprintf to log a message at TraceLevel if it's enabled.
// It also adds a prefix "[TRACE] " to the message.
//
// If trace messages are disabled globally, calling this function is
// a no-op.
func (s SugaredLogger) Tracef(format string, args ...any) {
	if globals.Level.Load() <= int32(TraceLevel) {
		s.slowPathTracef(format, args)
	}
}

func (s SugaredLogger) slowPathTracef(format string, args []any) {
	logger := s.SugaredLogger.Desugar().WithOptions(zap.AddCallerSkip(3))
	msg := formatMessage(format, args)
	checkAndWriteTraceMessage(logger, msg)
}

func checkAndWriteTraceMessage(l *zap.Logger, msg string, fields ...zap.Field) {
	if ce := l.Check(TraceLevel, msg); ce != nil {
		fileName := ce.Caller.File
		if fileName != "" {
			_, fileName, _, _, _ = getCaller(3)
		}
		if fileName != "" && globals.Props.traceFilter != nil &&
			!globals.Props.traceFilter.Allow(fileName) {
			return
		}
		ce.Write(fields...)
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
// If trace messages are disabled globally, calling this function is
// a no-op.
func TRACE(args ...any) {
	if globals.Level.Load() <= int32(TraceLevel) {
		_slowPathTrace(0, nil, args)
	}
}

// TRACESkip is similar to TRACE, but it has an extra skip argument to get
// correct caller information. When you need to wrap TRACE, you will always
// want to use this function instead of TRACE.
//
// If trace messages are disabled globally, calling this function is
// a no-op.
func TRACESkip(skip int, args ...any) {
	if globals.Level.Load() <= int32(TraceLevel) {
		_slowPathTrace(skip, nil, args)
	}
}

// TRACE1 is same with TRACE, but it accepts an extra arg0 before args.
func TRACE1(arg0 any, args ...any) {
	if globals.Level.Load() <= int32(TraceLevel) {
		_slowPathTrace(0, arg0, args)
	}
}

// TRACESkip1 is same with TRACESkip, but it accepts an extra arg0 before args.
func TRACESkip1(skip int, arg0 any, args ...any) {
	if globals.Level.Load() <= int32(TraceLevel) {
		_slowPathTrace(skip, arg0, args)
	}
}

func _slowPathTrace(skip int, a0 any, args []any) {
	caller, fullFileName, simpleFileName, line, _ := getCaller(skip + 2)
	if fullFileName != "" && globals.Props.traceFilter != nil &&
		!globals.Props.traceFilter.Allow(fullFileName) {
		return
	}
	logger, msg, fields := parseLoggerAndParams(skip, a0, args)
	if msg == "" {
		msg = fmt.Sprintf("========  %s#L%d - %s  ========", simpleFileName, line, caller)
	} else {
		msg = fmt.Sprintf("[%s] %s", caller, msg)
	}
	if ce := logger.Check(TraceLevel, msg); ce != nil {
		ce.Write(fields...)
	}
}

func parseLoggerAndParams(skip int, a0 any, args []any) (Logger, string, []zap.Field) {
	isArgs0 := false
	if a0 == nil && len(args) > 0 {
		a0 = args[0]
		isArgs0 = true
	}
	trimArg0 := func(args []any) []any {
		if isArgs0 {
			args = args[1:]
		}
		return args
	}
	var logger = L()
	if a0 != nil {
		switch a0 := a0.(type) {
		case context.Context:
			logger = WithCtx(a0)
			args = trimArg0(args)
		case Logger:
			logger = a0
			args = trimArg0(args)
		case SugaredLogger:
			logger = a0.Desugar()
			args = trimArg0(args)
		case *zap.Logger:
			logger = Logger{Logger: a0}
			args = trimArg0(args)
		case *zap.SugaredLogger:
			logger = Logger{Logger: a0.Desugar()}
			args = trimArg0(args)
		default:
			if !isArgs0 {
				args = append([]any{a0}, args...)
			}
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
	}
	return logger, formatMessage(template, args), nil
}

func tryConvertFields(args []any) ([]zap.Field, bool) {
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

func formatMessage(template string, fmtArgs []any) string {
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
	return fmt.Sprint(fmtArgs...)
}
