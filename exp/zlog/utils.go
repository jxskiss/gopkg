package zlog

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"
)

func TRACE(args ...interface{}) {
	if GetLevel() <= TraceLevel {
		logger, msg, fields := parseLoggerAndParams(args)
		msg = addCallerPrefix(0, "TRACE", msg)
		logger.Debug(msg, fields...)
	}
}

func TRACESkip(skip int, args ...interface{}) {
	if GetLevel() <= TraceLevel {
		logger, msg, fields := parseLoggerAndParams(args)
		msg = addCallerPrefix(skip, "TRACE", msg)
		logger.WithOptions(zap.AddCallerSkip(skip)).Debug(msg, fields...)
	}
}

func parseLoggerAndParams(args []interface{}) (*zap.Logger, string, []zap.Field) {
	if len(args) == 0 {
		return _l(), "", nil
	}

	var logger *zap.Logger
	switch arg0 := args[0].(type) {
	case context.Context:
		logger = B(arg0).Build().WithOptions(zap.AddCallerSkip(1))
		args = args[1:]
	case *zap.Logger:
		logger = arg0.WithOptions(zap.AddCallerSkip(1))
		args = args[1:]
	case *zap.SugaredLogger:
		logger = arg0.Desugar().WithOptions(zap.AddCallerSkip(1))
		args = args[1:]
	default:
		logger = _l()
	}
	if len(args) == 0 {
		return logger, "", nil
	}

	template := ""
	if s, ok := args[0].(string); ok && strings.IndexByte(s, '%') >= 0 {
		template = s
		args = args[1:]
		if len(args) == 0 {
			return logger, template, nil
		}
	}

	isZapFields := true
	for i := 0; i < len(args); i++ {
		if _, ok := args[i].(zap.Field); !ok {
			isZapFields = false
			break
		}
	}
	if isZapFields {
		return logger, template, convertFields(args)
	}
	return logger, formatMessage(template, args), nil
}

func addCallerPrefix(skip int, level, msg string) string {
	caller, file, line, _ := getCaller(skip + 2)
	if msg == "" {
		return fmt.Sprintf("[%s] ========  %s#L%d - %s  ========", level, file, line, caller)
	}
	return fmt.Sprintf("[%s] [%s] %s", level, caller, msg)
}

func convertFields(fields []interface{}) []zap.Field {
	if len(fields) == 0 {
		return nil
	}
	out := make([]zap.Field, len(fields))
	for i, f := range fields {
		out[i] = f.(zap.Field)
	}
	return out
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
