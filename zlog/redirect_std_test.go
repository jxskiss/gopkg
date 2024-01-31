package zlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectLevel(t *testing.T) {
	testData := []struct {
		Message   string
		WantLevel Level
		WantOk    bool
	}{
		{"[TRACE] ", TraceLevel, true},
		{"[Trace]", TraceLevel, true},
		{"[trace]some message", TraceLevel, true},
		{"[TRACE] some message", TraceLevel, true},
		{"trace some message", 0, false},
		{"trace: some message", TraceLevel, true},
		{"tracesomemessage", 0, false},

		{"[DEBUG] ", DebugLevel, true},
		{"[Debug] some message", DebugLevel, true},
		{"[debug]some message", DebugLevel, true},
		{"[DEBUG]some message", DebugLevel, true},
		{"DEBUG some message", 0, false},
		{"DEBUG: some message", DebugLevel, true},
		{"DEBUGsome message", 0, false},

		{"[INFO] ", InfoLevel, true},
		{"[Info]", InfoLevel, true},
		{"[info]some message", InfoLevel, true},
		{"info: some message", InfoLevel, true},
		{"INFO some message", 0, false},
		{"some info message", 0, false},

		{"[WARN] ", WarnLevel, true},
		{"[WARNING] ", WarnLevel, true},
		{"warn: some message", WarnLevel, true},
		{"[Warn] some message", WarnLevel, true},
		{"WARN some message", 0, false},
		{"warning message", 0, false},
		{"[Warning] some message", WarnLevel, true},
		{"[WARNING] message", WarnLevel, true},
		{"warning: message", WarnLevel, true},

		{"[ERROR] ", ErrorLevel, true},
		{"[Error]", ErrorLevel, true},
		{"[error]some message", ErrorLevel, true},
		{"[ERROR] some message", ErrorLevel, true},
		{"error some message", 0, false},
		{"error: some message", ErrorLevel, true},
		{"errormessage", 0, false},

		{"[PANIC] ", ErrorLevel, true},
		{"[Panic]", ErrorLevel, true},
		{"panic: some message", ErrorLevel, true},
		{"panic some message", 0, false},

		{"[FATAL] ", ErrorLevel, true},
		{"[Fatal]", ErrorLevel, true},
		{"Fatal: some message", ErrorLevel, true},
		{"Fatal error occurred", 0, false},
	}

	for _, tc := range testData {
		gotLevel, gotOk := detectLevel(tc.Message)
		assert.Equal(t, tc.WantLevel, gotLevel)
		assert.Equal(t, tc.WantOk, gotOk)
	}
}
