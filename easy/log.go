package easy

import (
	"bytes"
	"context"
	"fmt"
	"github.com/json-iterator/go"
	"github.com/jxskiss/gopkg/json"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"unicode/utf8"
)

func ConfigLog(
	enableDebug bool,
	defaultLogger ErrDebugLogger,
	ctxFunc func(ctx context.Context) ErrDebugLogger,
) {
	logcfg.EnableDebug = enableDebug
	if defaultLogger != nil {
		logcfg.DefaultLogger = defaultLogger
	}
	if ctxFunc != nil {
		logcfg.CtxFunc = ctxFunc
	}
}

var logcfg = struct {
	EnableDebug   bool
	DefaultLogger ErrDebugLogger
	// CtxFunc retrieves logger from context.Context.
	CtxFunc func(ctx context.Context) ErrDebugLogger
}{
	EnableDebug:   false,
	DefaultLogger: stdLogger{},
}

type stdLogger struct{}

func (p stdLogger) Debugf(format string, args ...interface{}) { log.Printf("DEBUG: "+format, args...) }
func (p stdLogger) Errorf(format string, args ...interface{}) { log.Printf("ERROR: "+format, args...) }

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

// ErrDebugLogger is an interface which log messages at ERROR and DEBUG level.
// It's implemented by *logrus.Logger, *logrus.Entry, *zap.SugaredLogger,
// and many other logging packages.
type ErrDebugLogger interface {
	ErrLogger
	DebugLogger
}

// PrintFunc is a function to print the given arguments in format to somewhere.
// It implements the interface `ErrDebugLogger`.
type PrintFunc func(format string, args ...interface{})

func (f PrintFunc) Errorf(format string, args ...interface{}) { f(format, args...) }

func (f PrintFunc) Debugf(format string, args ...interface{}) { f(format, args...) }

// Printer is an interface which writes log messages to somewhere.
// It's implemented by *logrus.Logger, *logrus.Entry, and many other
// logging packages.
type Printer interface {
	Printf(format string, args ...interface{})
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

// Pretty converts given object to a pretty formatted json string.
// If the input is an json string, it will be formatted using json.Indent
// with four space characters as indent.
func Pretty(v interface{}) string {
	switch v.(type) {
	case []byte, string:
		src := ToBytes_(v)
		if json.Valid(src) {
			buf := bytes.NewBuffer(nil)
			_ = json.Indent(buf, src, "", "    ")
			return String_(buf.Bytes())
		}
		if utf8.Valid(src) {
			return src.String_()
		}
		return "<pretty: non-printable bytes>"
	default:
		buf, err := logjson.MarshalIndent(v, "", "    ")
		if err != nil {
			return fmt.Sprintf("<error: %v>", err)
		}
		return String_(buf)
	}
}

// Caller returns function name, filename, and the line number of the caller.
// The argument skip is the number of stack frames to ascend, with 0
// identifying the caller of Caller.
func Caller(skip int) (name, file string, line int) {
	pc, file, line, _ := runtime.Caller(skip + 1)
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

var (
	stdoutMu sync.Mutex
	stdlogMu sync.Mutex
)

// CopyStdout replaces os.Stdout with a file created by `os.Pipe()`, and
// copies the content written to os.Stdout.
// This is not safe and most likely problematic, it's mainly to help intercepting
// output in testing.
func CopyStdout(f func()) (Bytes, error) {
	stdoutMu.Lock()
	defer stdoutMu.Unlock()
	old := os.Stdout
	defer func() { os.Stdout = old }()

	r, w, err := os.Pipe()
	// just to make sure the error didn't happen
	// in case of unfortunate, we should still do the specified work
	if err != nil {
		f()
		return nil, err
	}

	// copy the output in a separate goroutine, so printing can't block indefinitely
	outCh := make(chan []byte)
	go func() {
		var buf bytes.Buffer
		multi := io.MultiWriter(&buf, old)
		io.Copy(multi, r)
		outCh <- buf.Bytes()
	}()

	// do the work, write the stdout to pipe
	os.Stdout = w
	f()
	w.Close()

	out := <-outCh
	return out, nil
}

// CopyStdLog replaces the out Writer of the default logger of `log` package,
// and copies the content written to it.
// This is unsafe and most likely problematic, it's mainly to help intercepting
// log output in testing.
//
// Also NOTE if the out Writer of the default logger has already been replaced
// with another writer, we won't know anything about that writer and will
// restore the out Writer to os.Stderr before it returns.
// It will be a real mess.
func CopyStdLog(f func()) Bytes {
	stdlogMu.Lock()
	defer stdlogMu.Unlock()
	defer log.SetOutput(os.Stderr)

	var buf bytes.Buffer
	multi := io.MultiWriter(&buf, os.Stderr)
	log.SetOutput(multi)
	f()
	return buf.Bytes()
}
