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
		logger = logger.WithOptions(zap.AddCallerSkip(1))
		logger.Debug(tracePrefix+msg, fields...)
	}
}

func LTracef(logger *zap.SugaredLogger, format string, args ...interface{}) {
	if GetLevel() <= TraceLevel {
		logger = logger.Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar()
		logger.Debugf(tracePrefix+format, args...)
	}
}
