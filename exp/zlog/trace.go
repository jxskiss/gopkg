package zlog

import "go.uber.org/zap"

func Trace(msg string, fields ...zap.Field) {
	if GetLevel() <= TraceLevel {
		msg = "[Trace] " + msg
		_l().Debug(msg, fields...)
	}
}

func Tracef(format string, args ...interface{}) {
	if GetLevel() <= TraceLevel {
		format = "[Trace] " + format
		_s().Debugf(format, args...)
	}
}

func LTrace(logger *zap.Logger, msg string, fields ...zap.Field) {
	if GetLevel() <= TraceLevel {
		msg = "[Trace] " + msg
		logger.Debug(msg, fields...)
	}
}

func LTracef(logger *zap.SugaredLogger, format string, args ...interface{}) {
	if GetLevel() <= TraceLevel {
		format = "[Trace] " + format
		logger.Debugf(format, args...)
	}
}
