package easy

import (
	"fmt"
	"github.com/json-iterator/go"
	"log"
	"runtime"
)

// ErrLogger is an interface which log an message at ERROR level.
// It's implemented by *logrus.Logger, *logrus.Entry, *zap.SugaredLogger,
// and many other logging packages.
type ErrLogger interface {
	Errorf(format string, args ...interface{})
}

// DebugLogger is an interface which log an message at DEBUG level.
// It's implemented by *logrus.Logger, *logrus.Entry, *zap.SugaredLogger,
// and many other logging packages.
type DebugLogger interface {
	Debugf(format string, args ...interface{})
}

// PrintFunc is a function to print the given arguments in format to somewhere.
// It implements both `ErrLogger` and `DebugLogger`.
type PrintFunc func(format string, args ...interface{})

func (f PrintFunc) Errorf(format string, args ...interface{}) { f(format, args...) }

func (f PrintFunc) Debugf(format string, args ...interface{}) { f(format, args...) }

// Printer is an interface which writes log messages to somewhere.
// It's implemented by *logrus.Logger, *logrus.Entry, and many other
// logging packages.
type Printer interface {
	Printf(format string, args ...interface{})
}

func logError(logger interface{}, format string, args ...interface{}) {
	switch logger := logger.(type) {
	case ErrLogger:
		logger.Errorf(format, args...)
	case Printer:
		logger.Printf(format, args...)
	case PrintFunc:
		logger(format, args...)
	case func(string, ...interface{}):
		logger(format, args...)
	default:
		log.Printf(format, args...)
	}
}

var logjson = jsoniter.Config{
	// compatible with standard library behavior
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,

	// incompatible with standard library behavior
	EscapeHTML: false,
}.Froze()

// JSON converts given object to a json string, it never returns error.
func JSON(v interface{}) string {
	b, err := logjson.Marshal(v)
	if err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	return String_(b)
}

// Caller returns function name, filename, and the line number of the caller.
func Caller(skip int) (name, file string, line int) {
	pc, file, line, _ := runtime.Caller(skip)
	name = runtime.FuncForPC(pc).Name()
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '/' {
			name = name[i+1:]
			break
		}
	}
	pathSepCnt := 0
	for i := len(file) - 1; i >= 0; i-- {
		if file[i] == '/' {
			pathSepCnt++
			if pathSepCnt == 2 {
				file = file[i+1:]
				break
			}
		}
	}
	return
}
