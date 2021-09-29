package zlog

import "go.uber.org/zap"

const tracePrefix = "[Trace] "

func Trace(msg string, fields ...zap.Field) {
	if GetLevel() <= TraceLevel {
		msg = tracePrefix + msg
		_l().Debug(msg, fields...)
	}
}

func Tracef(format string, args ...interface{}) {
	if GetLevel() <= TraceLevel {
		format = tracePrefix + format
		_s().Debugf(format, args...)
	}
}

func LTrace(logger *zap.Logger, msg string, fields ...zap.Field) {
	if GetLevel() <= TraceLevel {
		msg = tracePrefix + msg
		logger.Debug(msg, fields...)
	}
}

func LTracef(logger *zap.SugaredLogger, format string, args ...interface{}) {
	if GetLevel() <= TraceLevel {
		format = tracePrefix + format
		logger.Debugf(format, args...)
	}
}
