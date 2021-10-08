package zlog

import "go.uber.org/zap"

func Trace(msg string, fields ...zap.Field) {
	if GetLevel() <= TraceLevel {
		_l().Debug(tracePrefix+msg, fields...)
	}
}

func Tracef(format string, args ...interface{}) {
	if GetLevel() <= TraceLevel {
		_s().Debugf(tracePrefix+format, args...)
	}
}

func LTrace(logger *zap.Logger, msg string, fields ...zap.Field) {
	if GetLevel() <= TraceLevel {
		logger.Debug(tracePrefix+msg, fields...)
	}
}

func LTracef(logger *zap.SugaredLogger, format string, args ...interface{}) {
	if GetLevel() <= TraceLevel {
		logger.Debugf(tracePrefix+format, args...)
	}
}
