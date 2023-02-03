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
		{TracePrefix, TraceLevel, true},
		{"[Trace]", TraceLevel, true},
		{"[trace]some message", TraceLevel, true},
		{"[TRACE] some message", TraceLevel, true},
		{"trace some message", TraceLevel, true},
		{"trace: some message", TraceLevel, true},
		{"tracesomemessage", 0, false},

		{DebugPrefix, DebugLevel, true},
		{"[Debug] some message", DebugLevel, true},
		{"[debug]some message", DebugLevel, true},
		{"[DEBUG]some message", DebugLevel, true},
		{"DEBUG some message", DebugLevel, true},
		{"DEBUG: some message", DebugLevel, true},
		{"DEBUGsome message", 0, false},

		{InfoPrefix, InfoLevel, true},
		{"[Info]", InfoLevel, true},
		{"[info]some message", InfoLevel, true},
		{"info: some message", InfoLevel, true},
		{"INFO some message", InfoLevel, true},
		{"some info message", 0, false},

		{NoticePrefix, NoticeLevel, true},
		{"[Notice] ", NoticeLevel, true},
		{"[notice]some message", NoticeLevel, true},
		{"NOTICE: some message", NoticeLevel, true},
		{"NOTICE some message", NoticeLevel, true},

		{WarnPrefix, WarnLevel, true},
		{"warn: some message", WarnLevel, true},
		{"[Warn] some message", WarnLevel, true},
		{"WARN some message", WarnLevel, true},
		{"warning message", WarnLevel, true},
		{"[Warning] some message", WarnLevel, true},
		{"[WARNING] message", WarnLevel, true},

		{ErrorPrefix, ErrorLevel, true},
		{"[Error]", ErrorLevel, true},
		{"[error]some message", ErrorLevel, true},
		{"[ERROR] some message", ErrorLevel, true},
		{"error some message", ErrorLevel, true},
		{"error: some message", ErrorLevel, true},
		{"errormessage", 0, false},

		{CriticalPrefix, CriticalLevel, true},
		{"[Critical]", CriticalLevel, true},
		{"critical some message", CriticalLevel, true},
		{"cRiTiCal: some message", CriticalLevel, true},

		{PanicPrefix, CriticalLevel, true},
		{"[Panic]", CriticalLevel, true},
		{"panic: some message", CriticalLevel, true},
		{"panic some message", CriticalLevel, true},

		{FatalPrefix, CriticalLevel, true},
		{"[Fatal]", CriticalLevel, true},
		{"Fatal: some message", CriticalLevel, true},
		{"Fatal error occurred", CriticalLevel, true},
	}

	for _, tc := range testData {
		gotLevel, gotOk := detectLevel(tc.Message)
		assert.Equal(t, tc.WantLevel, gotLevel)
		assert.Equal(t, tc.WantOk, gotOk)
	}
}
