package easy

import (
	"bytes"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"regexp"
	"testing"
)

func TestJSON(t *testing.T) {
	tests := []map[string]interface{}{
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
		x := JSON(test["value"])
		assert.Equal(t, test["want"], x)
	}
}

func TestCaller(t *testing.T) {
	name, file, line := Caller(1)
	assert.Equal(t, "easy.TestCaller", name)
	assert.Equal(t, "easy/log_test.go", file)
	assert.Equal(t, 40, line)
}

func TestDEBUG_bare_func(t *testing.T) {
	logbuf := bytes.NewBuffer(nil)
	log.SetOutput(logbuf)
	defer log.SetOutput(os.Stderr)

	// test func()
	ConfigDebugLog(true, nil, nil)
	msg := "test DEBUG_bare_func"
	DEBUG(func() {
		log.Println(msg, 1, 2, 3)
	})
	got := logbuf.String()
	assert.Contains(t, got, msg)
	assert.Contains(t, got, "1 2 3")
}

func TestDEBUG_logger_interface(t *testing.T) {
	logbuf := bytes.NewBuffer(nil)
	logger := &bufDebugLogger{buf: logbuf}

	// test logger interface
	ConfigDebugLog(true, nil, nil)
	msg := "test DEBUG_logger_interface"
	DEBUG(logger, msg, 1, 2, 3)
	got := logbuf.String()
	assert.Contains(t, got, msg)
	assert.Contains(t, got, "1 2 3")
}

func TestDEBUG_logger_func(t *testing.T) {
	logbuf := bytes.NewBuffer(nil)
	log.SetOutput(logbuf)
	defer log.SetOutput(os.Stderr)

	// test logger function
	ConfigDebugLog(true, nil, nil)
	msg := "test DEBUG_logger_func"
	logger := func() *stdLogger {
		return &stdLogger{}
	}
	DEBUG(logger, msg, 1, 2, 3)
	got := logbuf.String()
	assert.Contains(t, got, msg)
	assert.Contains(t, got, "1 2 3")
}

func TestDEBUG_print_func(t *testing.T) {
	logbuf := bytes.NewBuffer(nil)
	log.SetOutput(logbuf)
	defer log.SetOutput(os.Stderr)

	// test print function
	ConfigDebugLog(true, nil, nil)
	msg := "test DEBUG_print_func"
	prefix := "PREFIX: "
	logger := func(format string, args ...interface{}) {
		format = prefix + format
		log.Printf(format, args...)
	}
	DEBUG(logger, msg, 1, 2, 3)
	got := logbuf.String()
	assert.Contains(t, got, prefix)
	assert.Contains(t, got, msg)
	assert.Contains(t, got, "1 2 3")
}

func TestDEBUG_ctx_logger(t *testing.T) {
	logbuf := bytes.NewBuffer(nil)
	ctx := context.WithValue(context.Background(), "TEST_LOGGER", &bufDebugLogger{buf: logbuf})
	getCtxLogger := func(ctx context.Context) DebugLogger {
		return ctx.Value("TEST_LOGGER").(*bufDebugLogger)
	}

	// test ctx logger
	ConfigDebugLog(true, nil, getCtxLogger)
	msg := "test DEBUG_ctx_logger"
	DEBUG(ctx, msg, 1, 2, 3)
	got := logbuf.String()
	assert.Contains(t, got, msg)
	assert.Contains(t, got, "1 2 3")
}

func TestDEBUG_simple(t *testing.T) {
	logbuf := bytes.NewBuffer(nil)
	log.SetOutput(logbuf)
	defer log.SetOutput(os.Stderr)
	ConfigDebugLog(true, nil, nil)

	// format
	DEBUG("test DEBUG_simple a=%v b=%v c=%v", 1, 2, 3)
	want1 := "test DEBUG_simple a=1 b=2 c=3"
	got1 := logbuf.String()
	assert.Contains(t, got1, want1)

	// raw params
	logbuf.Reset()
	DEBUG("test DEBUG_simple a=", 1, "b=", 2, "c=", 3)
	want2 := "test DEBUG_simple a= 1 b= 2 c= 3"
	got2 := logbuf.String()
	assert.Contains(t, got2, want2)
}

func TestDEBUG_empty(t *testing.T) {
	logbuf := bytes.NewBuffer(nil)
	log.SetOutput(logbuf)
	defer log.SetOutput(os.Stderr)
	ConfigDebugLog(true, nil, nil)

	DEBUG()
	got := logbuf.String()
	want := regexp.MustCompile(`DEBUG: easy/log_test.go#L\d+ - easy.TestDEBUG_empty`)
	assert.Regexp(t, want, got)
}

func TestDEBUGSkip(t *testing.T) {
	logbuf := bytes.NewBuffer(nil)
	log.SetOutput(logbuf)
	defer log.SetOutput(os.Stderr)
	ConfigDebugLog(true, nil, nil)

	DEBUGWrap()
	got := logbuf.String()
	want := regexp.MustCompile(`DEBUG: easy/log_test.go#L\d+ - easy.TestDEBUGSkip`)
	assert.Regexp(t, want, got)

	logbuf.Reset()
	DEBUGWrapSkip2()
	got = logbuf.String()
	want = regexp.MustCompile(`DEBUG: easy/log_test.go#L\d+ - easy.TestDEBUGSkip`)
	assert.Regexp(t, want, got)
}

func TestConfigDebugLog(t *testing.T) {
	defaultLogger := &bufDebugLogger{}
	getCtxLogger := func(ctx context.Context) DebugLogger {
		return ctx.Value("TEST_LOGGER").(*bufDebugLogger)
	}
	ConfigDebugLog(true, defaultLogger, getCtxLogger)
}

type bufDebugLogger struct {
	buf *bytes.Buffer
}

func (p *bufDebugLogger) Debugf(format string, args ...interface{}) {
	fmt.Fprintf(p.buf, format, args...)
}

func DEBUGWrap(args ...interface{}) {
	DEBUGSkip(1, args...)
}

func DEBUGWrapSkip2(args ...interface{}) {
	skip2 := func(args ...interface{}) {
		DEBUGSkip(2, args...)
	}
	skip2(args...)
}
