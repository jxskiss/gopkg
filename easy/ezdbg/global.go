package ezdbg

var globalLogger = NewLogger("", &globalCfg)

// DEBUG ... see [Logger.DEBUG].
func DEBUG(args ...any) {
	globalLogger.DEBUGSkip(1, args...)
}

// DEBUGSkip ... see [Logger.DEBUGSkip].
func DEBUGSkip(skip int, args ...any) {
	globalLogger.DEBUGSkip(skip+1, args...)
}

// PRETTY ... see [Logger.PRETTY].
func PRETTY(args ...any) {
	globalLogger.PRETTYSkip(1, args...)
}

// PRETTYSkip ... see [Logger.PRETTYSkip].
func PRETTYSkip(skip int, args ...any) {
	globalLogger.PRETTYSkip(skip+1, args...)
}
