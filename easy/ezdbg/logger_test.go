package ezdbg

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jxskiss/gopkg/v2/easy"
)

type simple struct {
	A string
}

type comptyp struct {
	I32   int32
	I32_p *int32

	I64   int64
	I64_p *int64

	Str   string
	Str_p *string

	Simple   simple
	Simple_p *simple
}

func TestJSON(t *testing.T) {
	tests := []map[string]any{
		{
			"value": 123,
			"want":  "123",
		},
		{
			"value": "456",
			"want":  `"456"`,
		},
		{
			"value": simple{"ABC"},
			"want":  `{"A":"ABC"}`,
		},
		{
			"value": "<html></html>",
			"want":  `"<html></html>"`,
		},
	}
	for _, test := range tests {
		x := easy.JSON(test["value"])
		assert.Equal(t, test["want"], x)
	}
}

func TestDEBUG_bare_func(t *testing.T) {
	// test func()
	configTestLog(true, nil)
	msg := "test DEBUG_bare_func"
	got := copyStdLog(func() {
		DEBUG(func() {
			log.Println(msg, 1, 2, 3)
		})
	})
	assert.Contains(t, string(got), msg)
	assert.Contains(t, string(got), "1 2 3")
}

func TestDEBUG_logger_interface(t *testing.T) {
	logbuf := bytes.NewBuffer(nil)
	logger := &bufLogger{buf: logbuf}

	// test logger interface
	configTestLog(true, nil)
	msg := "test DEBUG_logger_interface"
	DEBUG(logger, msg, 1, 2, 3)
	got := logbuf.String()
	assert.Contains(t, got, msg)
	assert.Contains(t, got, "1 2 3")
}

func TestDEBUG_logger_func(t *testing.T) {
	// test logger function
	configTestLog(true, nil)
	logger := func() stdLogger {
		return stdLogger{}
	}
	msg := "test DEBUG_logger_func"
	got := copyStdLog(func() {
		DEBUG(logger, msg, 1, 2, 3)
	})
	assert.Contains(t, string(got), msg)
	assert.Contains(t, string(got), "1 2 3")
}

func TestDEBUG_print_func(t *testing.T) {
	// test print function
	configTestLog(true, nil)
	msg := "test DEBUG_print_func"
	prefix := "PREFIX: "
	logger := func(format string, args ...any) {
		format = prefix + format
		log.Printf(format, args...)
	}
	got := copyStdLog(func() {
		DEBUG(logger, msg, 1, 2, 3)
	})
	assert.Contains(t, string(got), prefix)
	assert.Contains(t, string(got), msg)
	assert.Contains(t, string(got), "1 2 3")
}

func TestDEBUG_ctx_logger(t *testing.T) {
	logbuf := bytes.NewBuffer(nil)
	ctx := context.WithValue(context.Background(), "TEST_LOGGER", &bufLogger{buf: logbuf})
	getCtxLogger := func(ctx context.Context) DebugLogger {
		return ctx.Value("TEST_LOGGER").(*bufLogger)
	}

	// test ctx logger
	configTestLog(true, getCtxLogger)
	msg := "test DEBUG_ctx_logger"
	DEBUG(ctx, msg, 1, 2, 3)
	got := logbuf.String()
	assert.Contains(t, got, msg)
	assert.Contains(t, got, "1 2 3")
}

func TestDEBUG_simple(t *testing.T) {
	configTestLog(true, nil)

	// test format
	got1 := copyStdLog(func() {
		DEBUG("test DEBUG_simple a=%v b=%v c=%v", 1, 2, 3)
	})
	want1 := "test DEBUG_simple a=1 b=2 c=3"
	assert.Contains(t, string(got1), want1)

	// raw params
	got2 := copyStdLog(func() {
		DEBUG("test DEBUG_simple a=", 1, "b=", 2, "c=", 3)
	})
	want2 := "test DEBUG_simple a= 1 b= 2 c= 3"
	assert.Contains(t, string(got2), want2)
}

func TestDEBUG_pointers(t *testing.T) {
	configTestLog(true, nil)

	got := copyStdLog(func() {
		var x = 1234
		var p1 = &x
		var p2 = &p1
		var p3 *int
		var p4 **int
		DEBUG(x, p1, p2, p3, p4)
	})
	assert.Contains(t, string(got), "[DEBUG] [ezdbg.TestDEBUG_pointers.func1] 1234 1234 1234 null null")
}

func TestDEBUG_empty(t *testing.T) {
	configTestLog(true, nil)
	got := copyStdLog(func() { DEBUG() })
	want := regexp.MustCompile(`ezdbg/logger_test.go#L\d+ - ezdbg.TestDEBUG_empty`)
	assert.Regexp(t, want, string(got))
}

func TestDEBUGSkip(t *testing.T) {
	configTestLog(true, nil)

	got := copyStdLog(func() { DEBUGWrap() })
	want := regexp.MustCompile(`ezdbg/logger_test.go#L\d+ - ezdbg.TestDEBUGSkip`)
	assert.Regexp(t, want, string(got))

	got = copyStdLog(func() { DEBUGWrapSkip2() })
	want = regexp.MustCompile(`ezdbg/logger_test.go#L\d+ - ezdbg.TestDEBUGSkip`)
	assert.Regexp(t, want, string(got))
}

func TestConfigLog(t *testing.T) {
	getCtxLogger := func(ctx context.Context) DebugLogger {
		return ctx.Value("TEST_LOGGER").(*bufLogger)
	}
	configTestLog(true, getCtxLogger)
	assert.NotNil(t, globalCfg.LoggerFunc)
}

func TestFilterRule(t *testing.T) {
	msg := "test FilterRule"

	testCases := []struct {
		name     string
		rule     string
		contains bool
	}{
		{name: "default", rule: "", contains: true},
		{name: "allow all", rule: "allow=all", contains: true},
		{name: "allow explicitly 1", rule: "allow=ezdbg/*.go", contains: true},
		{name: "allow explicitly 2", rule: "allow=easy/**", contains: true},
		{name: "not allowed explicitly", rule: "allow=confr/*,easy/*.go,easy/ezhttp/**", contains: false},
		{name: "deny all", rule: "deny=all", contains: false},
		{name: "deny explicitly", rule: "deny=ezdbg/*", contains: false},
		{name: "not denied explicitly", rule: "deny=confr/*,easy/ezhttp/**", contains: true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			configTestLogWithFilterRule(true, nil, tc.rule)
			got := copyStdLog(func() {
				DEBUG(msg)
			})
			if tc.contains {
				assert.Contains(t, string(got), msg)
			} else {
				assert.NotContains(t, string(got), msg)
			}
		})
	}
}

func TestLoggerName(t *testing.T) {
	logger := NewLogger("test-logger", &Config{
		EnableDebug: func(ctx context.Context) bool { return true },
	})
	assert.Equal(t, "test-logger", logger.Name())

	msg := "test LoggerName"
	got := copyStdLog(func() {
		logger.DEBUG(msg)
	})
	assert.Contains(t, string(got), "[logger=test-logger] ")
	assert.Contains(t, string(got), "[ezdbg.TestLoggerName.func2] ")
	assert.Contains(t, string(got), msg)
}

// -------- utilities -------- //

type bufLogger struct {
	buf *bytes.Buffer
}

func (p *bufLogger) Debugf(format string, args ...any) {
	if p.buf == nil {
		p.buf = bytes.NewBuffer(nil)
	}
	fmt.Fprintf(p.buf, format, args...)
}

func (p *bufLogger) Errorf(format string, args ...any) {
	if p.buf == nil {
		p.buf = bytes.NewBuffer(nil)
	}
	fmt.Fprintf(p.buf, format, args...)
}

func DEBUGWrap(args ...any) {
	DEBUGSkip(1, args...)
}

func DEBUGWrapSkip2(args ...any) {
	skip2 := func(args ...any) {
		DEBUGSkip(2, args...)
	}
	skip2(args...)
}

func configTestLog(
	enableDebug bool,
	ctxLogger func(context.Context) DebugLogger,
) {
	configTestLogWithFilterRule(enableDebug, ctxLogger, "")
}

func configTestLogWithFilterRule(
	enableDebug bool,
	ctxLogger func(context.Context) DebugLogger,
	filterRule string,
) {
	ConfigGlobal(Config{
		EnableDebug: func(_ context.Context) bool { return enableDebug },
		LoggerFunc:  ctxLogger,
		FilterRule:  filterRule,
	})
}

var (
	stdoutMu sync.Mutex
	stdlogMu sync.Mutex
)

// copyStdout replaces os.Stdout with a file created by `os.Pipe()`, and
// copies the content written to os.Stdout.
// This is not safe and most likely problematic, it's mainly to help intercepting
// output in testing.
func copyStdout(f func()) ([]byte, error) {
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

// copyStdLog replaces the out Writer of the default logger of `log` package,
// and copies the content written to it.
// This is unsafe and most likely problematic, it's mainly to help intercepting
// log output in testing.
//
// Also NOTE if the out Writer of the default logger has already been replaced
// with another writer, we won't know anything about that writer and will
// restore the out Writer to os.Stderr before it returns.
// It will be a real mess.
func copyStdLog(f func()) []byte {
	stdlogMu.Lock()
	defer stdlogMu.Unlock()
	defer log.SetOutput(os.Stderr)

	var buf bytes.Buffer
	multi := io.MultiWriter(&buf, os.Stderr)
	log.SetOutput(multi)
	f()
	return buf.Bytes()
}
