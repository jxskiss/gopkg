//nolint:staticcheck
package zlog

import (
	"fmt"
	"log/slog"
	"os"
)

// Debugf ...
// Deprecated: this function will be removed in the future, please use Debug instead.
func Debugf(format string, args ...any) {
	l := Default()
	if l.Enabled(nil, slog.LevelDebug) {
		msg := formatMessage(format, args)
		_log(nil, 0, l, slog.LevelDebug, msg, nil)
	}
}

// Infof ...
// Deprecated: this function will be removed in the future, please use Info instead.
func Infof(format string, args ...any) {
	l := Default()
	if l.Enabled(nil, slog.LevelInfo) {
		msg := formatMessage(format, args)
		_log(nil, 0, l, slog.LevelInfo, msg, nil)
	}
}

// Warnf ...
// Deprecated: this function will be removed in the future, please use Warn instead.
func Warnf(format string, args ...any) {
	l := Default()
	if l.Enabled(nil, slog.LevelWarn) {
		msg := formatMessage(format, args)
		_log(nil, 0, l, slog.LevelWarn, msg, nil)
	}
}

// Errorf ...
// Deprecated: this function will be removed in the future, please use Error instead.
func Errorf(format string, args ...any) {
	msg := formatMessage(format, args)
	_log(nil, 0, Default(), slog.LevelError, msg, nil)
}

// Fatalf is equivalent to Errorf() followed by a call to os.Exit(1).
// Deprecated: this function will be removed in the future, please use Fatal instead.
func Fatalf(format string, args ...any) {
	msg := formatMessage(format, args)
	_log(nil, 0, Default(), slog.LevelError, msg, nil)
	os.Exit(1)
}

// Print uses [fmt.Sprint] to log a message at InfoLevel, or level
// detected from args, if it's enabled.
// It has same signature with [log.Print].
//
// Deprecated: this function will be removed in the future, please use
// Logger or the leveled logging functions.
func Print(args ...any) {
	l := Default()
	level := slog.LevelInfo
	if len(args) > 0 {
		s, _ := args[0].(string)
		level = detectLevel(s)
	}
	if l.Enabled(nil, level) {
		msg := fmt.Sprint(args...)
		_log(nil, 0, l, level, msg, nil)
	}
}

// Printf uses [fmt.Sprintf] to log a message at InfoLevel, or level
// detected from args, if it's enabled.
// It has same signature with [log.Printf].
//
// Deprecated: this function will be removed in the future, please use
// Logger or the leveled logging functions.
func Printf(format string, args ...any) {
	l := Default()
	level := detectLevel(format)
	if l.Enabled(nil, level) {
		msg := formatMessage(format, args)
		_log(nil, 0, l, level, msg, nil)
	}
}

// Println uses [fmt.Sprintln] to log a message at InfoLevel, or level
// detected from args, if it's enabled.
// It has same signature with [log.Println].
//
// Deprecated: this function will be removed in the future, please use
// Logger or the leveled logging functions.
func Println(args ...any) {
	l := Default()
	level := slog.LevelInfo
	if len(args) > 0 {
		s, _ := args[0].(string)
		level = detectLevel(s)
	}
	if l.Enabled(nil, level) {
		msg := fmt.Sprintln(args...)
		_log(nil, 0, l, level, msg, nil)
	}
}

func formatMessage(format string, args []any) string {
	if len(args) == 0 {
		return format
	}
	return fmt.Sprintf(format, args...)
}
