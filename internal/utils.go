package internal

import (
	"fmt"
	"runtime"
	"strings"
)

// IdentifyPanic reports the panic location when a panic happens.
func IdentifyPanic(extraSkip int) string {
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

	n := runtime.Callers(3+extraSkip, pc[:])
	frames := runtime.CallersFrames(pc[:n])
	for {
		f, more := frames.Next()
		name, file, line = f.Function, f.File, f.Line
		if !more || !strings.HasPrefix(name, "runtime.") {
			break
		}
	}
	switch {
	case name != "":
		return fmt.Sprintf("%v:%v", name, line)
	case file != "":
		return fmt.Sprintf("%v:%v", file, line)
	}

	return fmt.Sprintf("pc:%x", pc)
}
