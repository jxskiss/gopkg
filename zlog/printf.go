//nolint:staticcheck
package zlog

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

// Debugf logs a message at level slog.LevelDebug.
// Arguments are handled in the manner of [fmt.Printf].
func Debugf(format string, args ...any) {
	l := Default()
	if l.Enabled(nil, slog.LevelDebug) {
		msg := formatMessage(format, args)
		_log(context.Background(), 0, l, slog.LevelDebug, msg, nil)
	}
}

// Infof logs a message at level slog.LevelInfo.
// Arguments are handled in the manner of [fmt.Printf].
func Infof(format string, args ...any) {
	l := Default()
	if l.Enabled(nil, slog.LevelInfo) {
		msg := formatMessage(format, args)
		_log(context.Background(), 0, l, slog.LevelInfo, msg, nil)
	}
}

// Warnf logs a message at level slog.LevelWarn.
// Arguments are handled in the manner of [fmt.Printf].
func Warnf(format string, args ...any) {
	l := Default()
	if l.Enabled(nil, slog.LevelWarn) {
		msg := formatMessage(format, args)
		_log(context.Background(), 0, l, slog.LevelWarn, msg, nil)
	}
}

// Errorf logs a message at level slog.LevelError.
// Arguments are handled in the manner of [fmt.Printf].
func Errorf(format string, args ...any) {
	msg := formatMessage(format, args)
	_log(context.Background(), 0, Default(), slog.LevelError, msg, nil)
}

// Fatalf is equivalent to Errorf() followed by a call to os.Exit(1).
func Fatalf(format string, args ...any) {
	msg := formatMessage(format, args)
	_log(context.Background(), 0, Default(), slog.LevelError, msg, nil)
	os.Exit(1)
}

// Print logs a message at level slog.LevelInfo, or level detected
// from args, if it's enabled.
// It has same signature with [log.Print],
// arguments are handled in the manner of [fmt.Print].
func Print(args ...any) {
	l := Default()
	level := slog.LevelInfo
	if len(args) > 0 {
		s, _ := args[0].(string)
		level = detectLevel(s)
	}
	if l.Enabled(nil, level) {
		msg := fmt.Sprint(args...)
		_log(context.Background(), 0, l, level, msg, nil)
	}
}

// Printf logs a message at level slog.LevelInfo, or level detected
// from args, if it's enabled.
// It has same signature with [log.Printf],
// arguments are handled in the manner of [fmt.Printf].
func Printf(format string, args ...any) {
	l := Default()
	level := detectLevel(format)
	if l.Enabled(nil, level) {
		msg := formatMessage(format, args)
		_log(context.Background(), 0, l, level, msg, nil)
	}
}

// Println logs a message at level slog.LevelInfo, or level detected
// from args, if it's enabled.
// It has same signature with [log.Println],
// arguments are handled in the manner of [fmt.Println].
func Println(args ...any) {
	l := Default()
	level := slog.LevelInfo
	if len(args) > 0 {
		s, _ := args[0].(string)
		level = detectLevel(s)
	}
	if l.Enabled(nil, level) {
		msg := fmt.Sprintln(args...)
		if len(msg) > 0 && msg[len(msg)-1] == '\n' {
			msg = msg[:len(msg)-1]
		}
		_log(context.Background(), 0, l, level, msg, nil)
	}
}

func formatMessage(format string, args []any) string {
	if len(args) == 0 {
		return format
	}
	return fmt.Sprintf(format, args...)
}
