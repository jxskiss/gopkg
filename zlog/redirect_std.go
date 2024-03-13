package zlog

import (
	"bytes"
	"log"
	"os"
	"strings"

	"go.uber.org/zap"
)

var replaceSlogDefault func(l *zap.Logger, disableCaller bool) func()

// redirectStdLog redirects output from the standard library's package-global
// logger to the supplied logger, it detects level from the logging messages,
// or use InfoLevel as default.
// Since zap already handles caller annotations, timestamps, etc.,
// it automatically disables the standard library's annotations and prefixing.
//
// It returns a function to restore the original prefix and flags and reset the
// standard library's output to os.Stderr.
func redirectStdLog(l *zap.Logger, disableCaller bool) func() {
	resetPkgSlog := func() {}
	if replaceSlogDefault != nil {
		resetPkgSlog = replaceSlogDefault(l, disableCaller)
	}
	resetPkgLog := replaceLogDefault(l)
	return func() {
		resetPkgSlog()
		resetPkgLog()
	}
}

func replaceLogDefault(l *zap.Logger) func() {
	oldFlag := log.Flags()
	oldPrefix := log.Prefix()
	log.SetFlags(0)
	log.SetPrefix("")
	log.SetOutput((*stdLogWriter)(
		l.Named("stdlog").WithOptions(zap.AddCallerSkip(3))))
	return func() {
		log.SetFlags(oldFlag)
		log.SetPrefix(oldPrefix)
		log.SetOutput(os.Stderr)
	}
}

type stdLogWriter zap.Logger

func (l *stdLogWriter) Write(p []byte) (int, error) {
	n := len(p)
	p = bytes.TrimSpace(p)
	str := string(p)
	level, ok := detectLevel(str)
	if !ok {
		level = InfoLevel
	}
	(*zap.Logger)(l).Log(level, str)
	return n, nil
}

// detectLevel guess logging level by checking the begging of a message.
// Note that we don't guess level greater than ErrorLevel
// from this function to avoid crashing a program accidentally.
func detectLevel(message string) (Level, bool) {
	const levelPrefixMinLen = 5
	if len(message) < levelPrefixMinLen {
		return 0, false
	}
	end := uint8(':')
	if message[0] == '[' {
		end, message = ']', message[1:]
	}
	switch message[0] {
	case 'T', 't':
		if len(message) > 5 && message[5] == end &&
			strings.EqualFold("trace", message[:5]) {
			return TraceLevel, true
		}
	case 'D', 'd':
		if len(message) > 5 && message[5] == end &&
			strings.EqualFold("debug", message[:5]) {
			return DebugLevel, true
		}
	case 'I', 'i':
		if len(message) > 4 && message[4] == end &&
			strings.EqualFold("info", message[:4]) {
			return InfoLevel, true
		}
	case 'W', 'w':
		if len(message) > 4 && message[4] == end &&
			strings.EqualFold("warn", message[:4]) {
			return WarnLevel, true
		}
		if len(message) > 7 && message[7] == end &&
			strings.EqualFold("warning", message[:7]) {
			return WarnLevel, true
		}
	case 'E', 'e':
		if len(message) > 5 && message[5] == end &&
			strings.EqualFold("error", message[:5]) {
			return ErrorLevel, true
		}
	case 'P', 'p':
		if len(message) > 5 && message[5] == end &&
			strings.EqualFold("panic", message[:5]) {
			return ErrorLevel, true
		}
	case 'F', 'f':
		if len(message) > 5 && message[5] == end &&
			strings.EqualFold("fatal", message[:5]) {
			return ErrorLevel, true
		}
	}
	return 0, false
}
