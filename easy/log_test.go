package easy

import (
	"bytes"
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"math/rand"
	"regexp"
	"strings"
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

var prettyTestWant = strings.TrimSpace(`
{
    "1": 123,
    "b": "<html>"
}`)

func TestPretty(t *testing.T) {
	test := map[interface{}]interface{}{
		1:   123,
		"b": "<html>",
	}
	jsonString := JSON(test)
	assert.Equal(t, `{"1":123,"b":"<html>"}`, jsonString)

	got1 := Pretty(test)
	assert.Equal(t, prettyTestWant, got1)

	got2 := Pretty(jsonString)
	assert.Equal(t, prettyTestWant, got2)

	test3 := []byte("<fff> not a json object")
	got3 := Pretty(test3)
	assert.Equal(t, string(test3), got3)

	test4 := make([]byte, 16)
	rand.Read(test4)
	got4 := Pretty(test4)
	assert.Equal(t, "<pretty: non-printable bytes>", got4)
}

func TestCaller(t *testing.T) {
	name, file, line := Caller(0)
	assert.Equal(t, "easy.TestCaller", name)
	assert.Equal(t, "easy/log_test.go", file)
	assert.Equal(t, 71, line)
}

func TestCopyStdout(t *testing.T) {
	msg := "test CopyStdout"
	got, _ := CopyStdout(func() {
		fmt.Println(msg)
	})
	assert.Contains(t, got.String_(), msg)
}

func TestCopyStdLog(t *testing.T) {
	msg := "test CopyStdLog"
	got := CopyStdLog(func() {
		log.Println(msg)
	})
	assert.Contains(t, got.String_(), msg)
}

func TestDEBUG_bare_func(t *testing.T) {
	// test func()
	ConfigDebugLog(true, nil, nil)
	msg := "test DEBUG_bare_func"
	got := CopyStdLog(func() {
		DEBUG(func() {
			log.Println(msg, 1, 2, 3)
		})
	}).String_()
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
	// test logger function
	ConfigDebugLog(true, nil, nil)
	logger := func() *stdLogger {
		return &stdLogger{}
	}
	msg := "test DEBUG_logger_func"
	got := CopyStdLog(func() {
		DEBUG(logger, msg, 1, 2, 3)
	}).String_()
	assert.Contains(t, got, msg)
	assert.Contains(t, got, "1 2 3")
}

func TestDEBUG_print_func(t *testing.T) {
	// test print function
	ConfigDebugLog(true, nil, nil)
	msg := "test DEBUG_print_func"
	prefix := "PREFIX: "
	logger := func(format string, args ...interface{}) {
		format = prefix + format
		log.Printf(format, args...)
	}
	got := CopyStdLog(func() {
		DEBUG(logger, msg, 1, 2, 3)
	}).String_()
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
	ConfigDebugLog(true, nil, nil)

	// test format
	got1 := CopyStdLog(func() {
		DEBUG("test DEBUG_simple a=%v b=%v c=%v", 1, 2, 3)
	}).String_()
	want1 := "test DEBUG_simple a=1 b=2 c=3"
	assert.Contains(t, got1, want1)

	// raw params
	got2 := CopyStdLog(func() {
		DEBUG("test DEBUG_simple a=", 1, "b=", 2, "c=", 3)
	}).String_()
	want2 := "test DEBUG_simple a= 1 b= 2 c= 3"
	assert.Contains(t, got2, want2)
}

func TestDEBUG_empty(t *testing.T) {
	ConfigDebugLog(true, nil, nil)
	got := CopyStdLog(func() { DEBUG() }).String_()
	want := regexp.MustCompile(`DEBUG: easy/log_test.go#L\d+ - easy.TestDEBUG_empty`)
	assert.Regexp(t, want, got)
}

func TestDEBUGSkip(t *testing.T) {
	ConfigDebugLog(true, nil, nil)

	got := CopyStdLog(func() { DEBUGWrap() }).String_()
	want := regexp.MustCompile(`DEBUG: easy/log_test.go#L\d+ - easy.TestDEBUGSkip`)
	assert.Regexp(t, want, got)

	got = CopyStdLog(func() { DEBUGWrapSkip2() }).String_()
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
