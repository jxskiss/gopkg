package easy

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/jxskiss/gopkg/ptr"
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

func TestCaller(t *testing.T) {
	name, file, line := Caller(0)
	assert.Equal(t, "easy.TestCaller", name)
	assert.Equal(t, "easy/log_test.go", file)
	assert.Equal(t, 37, line)
}

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

func TestLogfmt(t *testing.T) {
	tests := []map[string]interface{}{
		{
			"value": 123,
			"want":  "123",
		},
		{
			"value": (*string)(nil),
			"want":  "null",
		},
		{
			"value": comptyp{
				I32:      32,
				I32_p:    ptr.Int32(32),
				I64:      64,
				I64_p:    nil,
				Str:      "str",
				Str_p:    ptr.String("str with space"),
				Simple:   simple{A: "simple.A"},
				Simple_p: nil,
			},
			"want": `i32=32 i32_p=32 i64=64 str=str str_p="str with space"`,
		},
		{
			"value": map[string]interface{}{
				"a": 1234,
				"b": "bcde",
				"c": 123.456,
				"d": simple{A: "simple.A"},
				"e": nil,
				"f": []byte("I'm bytes"),
			},
			"want": `a=1234 b=bcde c=123.456 f="I'm bytes"`,
		},
	}
	for _, test := range tests {
		got := Logfmt(test["value"])
		assert.Equal(t, test["want"], got)
	}
}

var prettyTestWant = strings.TrimSpace(`
{
    "1": 123,
    "b": "<html>"
}`)

func TestPretty(t *testing.T) {
	test := map[string]interface{}{
		"1": 123,
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

func TestCopyStdout(t *testing.T) {
	msg := "test CopyStdout"
	got, _ := CopyStdout(func() {
		fmt.Println(msg)
	})
	assert.Contains(t, string(got), msg)
}

func TestCopyStdLog(t *testing.T) {
	msg := "test CopyStdLog"
	got := CopyStdLog(func() {
		log.Println(msg)
	})
	assert.Contains(t, string(got), msg)
}

func TestDEBUG_bare_func(t *testing.T) {
	// test func()
	configTestLog(true, nil, nil)
	msg := "test DEBUG_bare_func"
	got := CopyStdLog(func() {
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
	configTestLog(true, nil, nil)
	msg := "test DEBUG_logger_interface"
	DEBUG(logger, msg, 1, 2, 3)
	got := logbuf.String()
	assert.Contains(t, got, msg)
	assert.Contains(t, got, "1 2 3")
}

func TestDEBUG_logger_func(t *testing.T) {
	// test logger function
	configTestLog(true, nil, nil)
	logger := func() stdLogger {
		return stdLogger{}
	}
	msg := "test DEBUG_logger_func"
	got := CopyStdLog(func() {
		DEBUG(logger, msg, 1, 2, 3)
	})
	assert.Contains(t, string(got), msg)
	assert.Contains(t, string(got), "1 2 3")
}

func TestDEBUG_print_func(t *testing.T) {
	// test print function
	configTestLog(true, nil, nil)
	msg := "test DEBUG_print_func"
	prefix := "PREFIX: "
	logger := func(format string, args ...interface{}) {
		format = prefix + format
		log.Printf(format, args...)
	}
	got := CopyStdLog(func() {
		DEBUG(logger, msg, 1, 2, 3)
	})
	assert.Contains(t, string(got), prefix)
	assert.Contains(t, string(got), msg)
	assert.Contains(t, string(got), "1 2 3")
}

func TestDEBUG_ctx_logger(t *testing.T) {
	logbuf := bytes.NewBuffer(nil)
	ctx := context.WithValue(context.Background(), "TEST_LOGGER", &bufLogger{buf: logbuf})
	getCtxLogger := func(ctx context.Context) ErrDebugLogger {
		return ctx.Value("TEST_LOGGER").(*bufLogger)
	}

	// test ctx logger
	configTestLog(true, nil, getCtxLogger)
	msg := "test DEBUG_ctx_logger"
	DEBUG(ctx, msg, 1, 2, 3)
	got := logbuf.String()
	assert.Contains(t, got, msg)
	assert.Contains(t, got, "1 2 3")
}

func TestDEBUG_simple(t *testing.T) {
	configTestLog(true, nil, nil)

	// test format
	got1 := CopyStdLog(func() {
		DEBUG("test DEBUG_simple a=%v b=%v c=%v", 1, 2, 3)
	})
	want1 := "test DEBUG_simple a=1 b=2 c=3"
	assert.Contains(t, string(got1), want1)

	// raw params
	got2 := CopyStdLog(func() {
		DEBUG("test DEBUG_simple a=", 1, "b=", 2, "c=", 3)
	})
	want2 := "test DEBUG_simple a= 1 b= 2 c= 3"
	assert.Contains(t, string(got2), want2)
}

func TestDEBUG_empty(t *testing.T) {
	configTestLog(true, nil, nil)
	got := CopyStdLog(func() { DEBUG() })
	want := regexp.MustCompile(`easy/log_test.go#L\d+ - easy.TestDEBUG_empty`)
	assert.Regexp(t, want, string(got))
}

func TestDEBUGSkip(t *testing.T) {
	configTestLog(true, nil, nil)

	got := CopyStdLog(func() { DEBUGWrap() })
	want := regexp.MustCompile(`easy/log_test.go#L\d+ - easy.TestDEBUGSkip`)
	assert.Regexp(t, want, string(got))

	got = CopyStdLog(func() { DEBUGWrapSkip2() })
	want = regexp.MustCompile(`easy/log_test.go#L\d+ - easy.TestDEBUGSkip`)
	assert.Regexp(t, want, string(got))
}

func TestConfigLog(t *testing.T) {
	defaultLogger := &bufLogger{}
	getCtxLogger := func(ctx context.Context) ErrDebugLogger {
		return ctx.Value("TEST_LOGGER").(*bufLogger)
	}
	configTestLog(true, defaultLogger, getCtxLogger)
}

type bufLogger struct {
	buf *bytes.Buffer
}

func (p *bufLogger) Debugf(format string, args ...interface{}) {
	if p.buf == nil {
		p.buf = bytes.NewBuffer(nil)
	}
	fmt.Fprintf(p.buf, format, args...)
}

func (p *bufLogger) Errorf(format string, args ...interface{}) {
	if p.buf == nil {
		p.buf = bytes.NewBuffer(nil)
	}
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

func configTestLog(
	enableDebug bool,
	defaultLogger ErrDebugLogger,
	ctxFunc func(ctx context.Context) ErrDebugLogger,
) {
	ConfigLog(LogCfg{
		EnableDebug: func() bool { return enableDebug },
		Logger:      func() ErrDebugLogger { return defaultLogger },
		CtxLogger:   ctxFunc,
	})
}
