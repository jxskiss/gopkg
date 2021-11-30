package zlog

import "testing"

func TestStdLogger(t *testing.T) {

	// By default, StdLogger is configured at info level (as the global
	// loggers), so this debug message won't be logged.
	// You may use SetupGlobals to configure the global loggers.
	StdLogger.Debugf("some debug message, %v, %v", "abc", "123")

	// The following messages will be logged by default.
	StdLogger.Infof("some info message, %v, %v", "abc", "123")
	StdLogger.Warnf("some warn message, %v, %v", "abc", "123")
	StdLogger.Errorf("some error message, %v, %v", "abc", "123")

	// Fatalf will log the message and exit the program.
	// StdLogger.Fatalf("some fatal message, %v, %v", "abc", "123")
}
