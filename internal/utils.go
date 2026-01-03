package internal

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"strings"
	"time"
)

func DefaultLoggerInfof(format string, args ...interface{}) {
	msg := format
	if len(args) > 0 {
		msg = fmt.Sprintf(format, args...)
	}
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:])
	r := slog.NewRecord(time.Now(), slog.LevelInfo, msg, pcs[0])
	_ = slog.Default().Handler().Handle(context.Background(), r)
}

func DefaultLoggerError(ctx context.Context, err error, msg string) {
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:])
	r := slog.NewRecord(time.Now(), slog.LevelError, msg, pcs[0])
	r.AddAttrs(slog.Any("error", err))
	_ = slog.Default().Handler().Handle(ctx, r)
}

// IdentifyPanic reports the panic location when a panic happens.
func IdentifyPanic(skip int) (location string, frames []runtime.Frame) {
	var name, file string
	var line int
	var pc [16]uintptr

	// Don't use runtime.FuncForPC here, it may give incorrect line number.
	//
	// From runtime.Callers' doc:
	//
	// To translate these PCs into symbolic information such as function
	// names and line numbers, use CallersFrames. CallersFrames accounts
	// for inlined functions and adjusts the return program counters into
	// call program counters. Iterating over the returned slice of PCs
	// directly is discouraged, as is using FuncForPC on any of the
	// returned PCs, since these cannot account for inlining or return
	// program counter adjustment.

	n := runtime.Callers(3+skip, pc[:])
	if n > 0 {
		callerFrames := runtime.CallersFrames(pc[:n])
		foundLoc := false
		frames = make([]runtime.Frame, 0, n)
		for {
			f, more := callerFrames.Next()
			frames = append(frames, f)
			if !foundLoc {
				name, file, line = f.Function, f.File, f.Line
				foundLoc = !strings.HasPrefix(name, "runtime.")
			}
			if !more {
				break
			}
		}
	}
	switch {
	case name != "":
		location = fmt.Sprintf("%v:%v", name, line)
	case file != "":
		location = fmt.Sprintf("%v:%v", file, line)
	default:
		location = fmt.Sprintf("pc:%x", pc)
	}
	return location, frames
}

func FormatFrames(frames []runtime.Frame) []byte {
	var buf bytes.Buffer
	for _, f := range frames {
		file, line, funcName := f.File, f.Line, f.Function
		if file == "" {
			file = "unknown"
		}
		if funcName == "" {
			funcName = "unknown"
		}
		if buf.Len() > 0 {
			buf.WriteByte('\n')
		}
		fmt.Fprintf(&buf, "%s:%d  (%s)", file, line, funcName)
	}
	return buf.Bytes()
}
