package easy

import (
	"fmt"
	"github.com/json-iterator/go"
	"log"
	"runtime"
)

type ErrLogger interface {
	Errorf(format string, args ...interface{})
}

type DebugLogger interface {
	Debugf(format string, args ...interface{})
}

type printfunc func(format string, args ...interface{})

type Printer interface {
	Printf(format string, args ...interface{})
}

func logError(logger interface{}, format string, args ...interface{}) {
	switch logger := logger.(type) {
	case ErrLogger:
		logger.Errorf(format, args...)
	case Printer:
		logger.Printf(format, args...)
	case printfunc:
		logger(format, args...)
	default:
		log.Printf(format, args...)
	}
}

var _json = jsoniter.Config{
	// compatible with standard library behavior
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,

	// incompatible with standard library behavior
	EscapeHTML: false,
}.Froze()

// JSON converts given object to a json string, it never returns error.
func JSON(v interface{}) string {
	b, err := _json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("<error: %v>", err)
	}
	return String_(b)
}

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
