// +build release

package easy

import "context"

func ConfigDebugLog(
	enableDebug bool,
	defaultLogger DebugLogger,
	ctxLoggerFunc func(context.Context) DebugLogger,
) {
}

func DEBUG(args ...interface{}) {}

func DEBUGSkip(skip int, args ...interface{}) {}

func SPEW(args ...interface{}) {}

func DUMP(args ...interface{}) {}
