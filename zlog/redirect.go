package zlog

import (
	"bytes"
	"context"
	"log"
	"log/slog"
	"strings"
)

func RedirectStdLog(l *Logger, attrs []slog.Attr) {
	log.SetFlags(0)
	log.SetPrefix("")
	log.SetOutput(&stdLogWriter{logger: l, attrs: attrs})
}

var defaultStdLogAttrs = []slog.Attr{slog.String("_logger", "stdlog")}

type stdLogWriter struct {
	logger *Logger
	attrs  []slog.Attr
}

func (l *stdLogWriter) Write(p []byte) (int, error) {
	logger := l.logger
	if logger == nil {
		logger = slog.Default()
	}
	n := len(p)
	p = bytes.TrimSpace(p)
	str := string(p)
	level := detectLevel(str)
	_logAttrs(2, context.Background(), logger, level, str, l.attrs)
	return n, nil
}

// detectLevel guess logging level by checking the begging of a message.
// Note that we don't guess level greater than ErrorLevel
// from this function to avoid crashing a program accidentally.
func detectLevel(message string) slog.Level {
	const levelPrefixMinLen = 5
	if len(message) < levelPrefixMinLen {
		return slog.LevelInfo
	}
	end := uint8(':')
	if message[0] == '[' {
		end, message = ']', message[1:]
	}
	switch message[0] {
	case 'T', 't':
		if len(message) > 5 && message[5] == end &&
			strings.EqualFold("trace", message[:5]) {
			return slog.LevelDebug
		}
	case 'D', 'd':
		if len(message) > 5 && message[5] == end &&
			strings.EqualFold("debug", message[:5]) {
			return slog.LevelDebug
		}
	case 'I', 'i':
		if len(message) > 4 && message[4] == end &&
			strings.EqualFold("info", message[:4]) {
			return slog.LevelInfo
		}
	case 'W', 'w':
		if len(message) > 4 && message[4] == end &&
			strings.EqualFold("warn", message[:4]) {
			return slog.LevelWarn
		}
		if len(message) > 7 && message[7] == end &&
			strings.EqualFold("warning", message[:7]) {
			return slog.LevelWarn
		}
	case 'E', 'e':
		if len(message) > 5 && message[5] == end &&
			strings.EqualFold("error", message[:5]) {
			return slog.LevelError
		}
	case 'P', 'p':
		if len(message) > 5 && message[5] == end &&
			strings.EqualFold("panic", message[:5]) {
			return slog.LevelError
		}
	case 'F', 'f':
		if len(message) > 5 && message[5] == end &&
			strings.EqualFold("fatal", message[:5]) {
			return slog.LevelError
		}
	}
	return slog.LevelInfo
}
