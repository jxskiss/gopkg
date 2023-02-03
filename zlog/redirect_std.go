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
	if message[0] == '[' {
		switch message[1] {
		case 'T', 't':
			if len(message[1:]) > 5 && message[6] == ']' &&
				strings.EqualFold("trace", message[1:6]) {
				return TraceLevel, true
			}
		case 'D', 'd':
			if len(message[1:]) > 5 && message[6] == ']' &&
				strings.EqualFold("debug", message[1:6]) {
				return DebugLevel, true
			}
		case 'I', 'i':
			if len(message[1:]) > 4 && message[5] == ']' &&
				strings.EqualFold("info", message[1:5]) {
				return InfoLevel, true
			}
		case 'N', 'n':
			if len(message[1:]) > 6 && message[7] == ']' &&
				strings.EqualFold("notice", message[1:7]) {
				return NoticeLevel, true
			}
		case 'W', 'w':
			if len(message[1:]) > 4 && message[5] == ']' &&
				strings.EqualFold("warn", message[1:5]) {
				return WarnLevel, true
			}
			if len(message[1:]) > 7 && message[8] == ']' &&
				strings.EqualFold("warning", message[1:8]) {
				return WarnLevel, true
			}
		case 'E', 'e':
			if len(message[1:]) > 5 && message[6] == ']' &&
				strings.EqualFold("error", message[1:6]) {
				return ErrorLevel, true
			}
		case 'C', 'c':
			if len(message[1:]) > 8 && message[9] == ']' &&
				strings.EqualFold("critical", message[1:9]) {
				return CriticalLevel, true
			}
		case 'P', 'p':
			if len(message[1:]) > 5 && message[6] == ']' &&
				strings.EqualFold("panic", message[1:6]) {
				return CriticalLevel, true
			}
		case 'F', 'f':
			if len(message[1:]) > 5 && message[6] == ']' &&
				strings.EqualFold("fatal", message[1:6]) {
				return CriticalLevel, true
			}
		}
	} else {
		switch message[0] {
		case 'T', 't':
			if len(message) > 5 && message[5] == ':' &&
				strings.EqualFold("trace", message[:5]) {
				return TraceLevel, true
			}
		case 'D', 'd':
			if len(message) > 5 && message[5] == ':' &&
				strings.EqualFold("debug", message[:5]) {
				return DebugLevel, true
			}
		case 'I', 'i':
			if len(message) > 4 && message[4] == ':' &&
				strings.EqualFold("info", message[:4]) {
				return InfoLevel, true
			}
		case 'N', 'n':
			if len(message) > 6 && message[6] == ':' &&
				strings.EqualFold("notice", message[:6]) {
				return NoticeLevel, true
			}
		case 'W', 'w':
			if len(message) > 4 && message[4] == ':' &&
				strings.EqualFold("warn", message[:4]) {
				return WarnLevel, true
			}
			if len(message) > 7 && message[7] == ':' &&
				strings.EqualFold("warning", message[:7]) {
				return WarnLevel, true
			}
		case 'E', 'e':
			if len(message) > 5 && message[5] == ':' &&
				strings.EqualFold("error", message[:5]) {
				return ErrorLevel, true
			}
		case 'C', 'c':
			if len(message) > 8 && message[8] == ':' &&
				strings.EqualFold("critical", message[:8]) {
				return CriticalLevel, true
			}
		case 'P', 'p':
			if len(message) > 5 && message[5] == ':' &&
				strings.EqualFold("panic", message[:5]) {
				return CriticalLevel, true
			}
		case 'F', 'f':
			if len(message) > 5 && message[5] == ':' &&
				strings.EqualFold("fatal", message[:5]) {
				return CriticalLevel, true
			}
		}
	}
	return 0, false
}
