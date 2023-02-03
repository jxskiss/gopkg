package zlog

import (
	"bytes"
	"log"
	"os"
	"strings"

	"go.uber.org/zap"
)

// RedirectStdLog redirects output from the standard library's package-global
// logger to the supplied logger, it detects level from the logging messages,
// or use InfoLevel as default.
// Since zap already handles caller annotations, timestamps, etc.,
// it automatically disables the standard library's annotations and prefixing.
//
// It returns a function to restore the original prefix and flags and reset the
// standard library's output to os.Stderr.
func RedirectStdLog(l *zap.Logger) func() {
	flags := log.Flags()
	prefix := log.Prefix()
	log.SetFlags(0)
	log.SetPrefix("")
	logger := l.WithOptions(zap.AddCallerSkip(3))
	log.SetOutput(&stdLogWriter{logger, InfoLevel})
	return func() {
		log.SetFlags(flags)
		log.SetPrefix(prefix)
		log.SetOutput(os.Stderr)
	}
}

type stdLogWriter struct {
	l     *zap.Logger
	level Level
}

func (l *stdLogWriter) Write(p []byte) (int, error) {
	n := len(p)
	p = bytes.TrimSpace(p)
	str := string(p)
	level, ok := detectLevel(str)
	if !ok {
		level = l.level
	}
	l.l.Log(level.ToZapLevel(), str)
	return n, nil
}

// Note that we don't guess level greater than CriticalLevel
// from this function to avoid crashing a program accidentally.
func detectLevel(message string) (Level, bool) {
	if len(message) < levelPrefixMinLen {
		return 0, false
	}
	i := 0
	if message[0] == '[' {
		i = 1
	}
	switch message[i] {
	case 'T', 't':
		if len(message[i:]) > 5 && isPrefixSep(message[i+5]) &&
			strings.EqualFold("trace", message[i:i+5]) {
			return TraceLevel, true
		}
	case 'D', 'd':
		if len(message[i:]) > 5 && isPrefixSep(message[i+5]) &&
			strings.EqualFold("debug", message[i:i+5]) {
			return DebugLevel, true
		}
	case 'I', 'i':
		if len(message[i:]) > 4 && isPrefixSep(message[i+4]) &&
			strings.EqualFold("info", message[i:i+4]) {
			return InfoLevel, true
		}
	case 'N', 'n':
		if len(message[i:]) > 6 && isPrefixSep(message[i+6]) &&
			strings.EqualFold("notice", message[i:i+6]) {
			return NoticeLevel, true
		}
	case 'W', 'w':
		if len(message[i:]) > 4 && isPrefixSep(message[i+4]) &&
			strings.EqualFold("warn", message[i:i+4]) {
			return WarnLevel, true
		}
		if len(message[i:]) > 7 && isPrefixSep(message[i+7]) &&
			strings.EqualFold("warning", message[i:i+7]) {
			return WarnLevel, true
		}
	case 'E', 'e':
		if len(message[i:]) > 5 && isPrefixSep(message[i+5]) &&
			strings.EqualFold("error", message[i:i+5]) {
			return ErrorLevel, true
		}
	case 'C', 'c':
		if len(message[i:]) > 8 && isPrefixSep(message[i+8]) &&
			strings.EqualFold("critical", message[i:i+8]) {
			return CriticalLevel, true
		}
	case 'P', 'p':
		if len(message[i:]) > 5 && isPrefixSep(message[i+5]) &&
			strings.EqualFold("panic", message[i:i+5]) {
			return CriticalLevel, true
		}
	case 'F', 'f':
		if len(message[i:]) > 5 && isPrefixSep(message[i+5]) &&
			strings.EqualFold("fatal", message[i:i+5]) {
			return CriticalLevel, true
		}
	}
	return 0, false
}

func isPrefixSep(b byte) bool {
	return b == ']' || b == ':' || b == ' '
}
