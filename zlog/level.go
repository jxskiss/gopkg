package zlog

import (
	"strconv"
	"strings"

	"go.uber.org/zap/zapcore"
)

// Level is an alias type of zapcore.Level.
type Level = zapcore.Level

const (
	// TraceLevel logs are the most fine-grained information which helps
	// developer "tracing" their code.
	// Users can expect this level to be very verbose, you can use it,
	// for example, to annotate each step in an algorithm or each individual
	// calling with parameters in your program.
	// TraceLevel logs should be disabled in production.
	TraceLevel Level = -2

	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel = zapcore.DebugLevel
	// InfoLevel is the default logging priority.
	InfoLevel = zapcore.InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel = zapcore.WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel = zapcore.ErrorLevel
	// DPanicLevel logs are particularly important errors. In development the
	// logger panics after writing the message.
	DPanicLevel = zapcore.DPanicLevel
	// PanicLevel logs a message, then panics.
	PanicLevel = zapcore.PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel = zapcore.FatalLevel
)

func unmarshalLevel(l *Level, text string) bool {
	switch text {
	case "trace", "TRACE":
		*l = TraceLevel
	case "debug", "DEBUG":
		*l = DebugLevel
	case "info", "INFO", "": // make the zero value useful
		*l = InfoLevel
	case "warn", "warning", "WARN", "WARNING":
		*l = WarnLevel
	case "error", "ERROR":
		*l = ErrorLevel
	case "dpanic", "DPANIC":
		*l = DPanicLevel
	case "panic", "PANIC":
		*l = PanicLevel
	case "fatal", "FATAL":
		*l = FatalLevel
	default:
		str := text
		if (strings.HasPrefix(str, "Level(") || strings.HasPrefix(str, "LEVEL(")) &&
			strings.HasSuffix(str, ")") {
			str = str[6 : len(str)-1]
		}
		i, err := strconv.Atoi(str)
		if err != nil {
			return false
		}
		*l = Level(i)
	}
	return true
}

func encodeLevelLowercase(lv zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	if lv == TraceLevel {
		enc.AppendString("trace")
	} else {
		enc.AppendString(lv.String())
	}
}

func encodeLevelCapital(lv zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	if lv == TraceLevel {
		enc.AppendString("TRACE")
	} else {
		enc.AppendString(lv.CapitalString())
	}
}
