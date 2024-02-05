package zlog

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestTrace(t *testing.T) {
	buf := &zaptest.Buffer{}
	l, p, _ := NewWithOutput(&Config{Level: "trace", Format: "logfmt"}, buf)
	defer ReplaceGlobals(l, p)()

	L().Trace("trace message", zap.String("k1", "v1"))

	got := buf.String()
	assert.Contains(t, got, "level=trace")
	assert.Contains(t, got, "k1=v1")
	assert.Contains(t, got, "caller=zlog/trace_test.go:18")
	assert.Contains(t, got, "level=trace")
	assert.Contains(t, got, `msg="trace message"`)
}

func TestTracef(t *testing.T) {
	buf := &zaptest.Buffer{}
	l, p, _ := NewWithOutput(&Config{Level: "trace", Format: "logfmt"}, buf)
	defer ReplaceGlobals(l, p)()

	S().Tracef("trace message, %v, %v", 123, 456)

	got := buf.String()
	assert.Contains(t, got, "level=trace")
	assert.Regexp(t, `caller=zlog/trace_test\.go:\d+`, got)
	assert.Contains(t, got, "level=trace")
	assert.Contains(t, got, `msg="trace message, 123, 456"`)
}

func TestTRACE(t *testing.T) {
	buf := &zaptest.Buffer{}
	l, p, _ := NewWithOutput(&Config{Level: "trace", Development: true}, buf)
	defer ReplaceGlobals(l, p)()

	TRACE()
	got := buf.String()
	assert.Regexp(t, `zlog/trace_test\.go:\d+`, got)

	buf.Reset()
	TRACE(WithLogger(context.Background(), L().With(zap.Int("TestTRACE", 1))))
	TRACE(L().With(zap.Int("TestTRACE", 1)))
	TRACE(S().With("TestTRACE", 1))
	for _, line := range buf.Lines() {
		assert.Contains(t, line, " TRACE ")
		assert.Regexp(t, `zlog/trace_test.go:5\d`, line)
		assert.Contains(t, line, `"TestTRACE": 1`)
	}

	buf.Reset()
	TRACE("a", "b", "c", 1, 2, 3)
	TRACE(context.Background(), "a", "b", "c", 1, 2, 3)
	TRACE(L(), "a", "b", "c", 1, 2, 3)
	TRACE(S(), "a", "b", "c", 1, 2, 3)
	for _, line := range buf.Lines() {
		assert.Contains(t, line, " TRACE ")
		assert.Regexp(t, `zlog/trace_test\.go:6\d`, line)
		assert.Contains(t, line, "abc1 2 3")
	}

	buf.Reset()
	TRACE("a=%v, b=%v, c=%v", 1, 2, 3)
	TRACE(context.Background(), "a=%v, b=%v, c=%v", 1, 2, 3)
	TRACE(L(), "a=%v, b=%v, c=%v", 1, 2, 3)
	TRACE(S(), "a=%v, b=%v, c=%v", 1, 2, 3)
	for _, line := range buf.Lines() {
		assert.Contains(t, line, " TRACE ")
		assert.Regexp(t, `zlog/trace_test\.go:7\d`, line)
		assert.Contains(t, line, "a=1, b=2, c=3")
	}
}

func TestTRACESkip(t *testing.T) {
	buf := &zaptest.Buffer{}
	l, p, _ := NewWithOutput(&Config{Level: "trace"}, buf)
	defer ReplaceGlobals(l, p)()

	TRACE()
	TRACESkip(0)
	wrappedTRACE() // this line outputs two messages

	lines := buf.Lines()
	assert.Len(t, lines, 4)
	for _, line := range lines {
		assert.Contains(t, line, `"level":"trace"`)
		assert.Contains(t, line, "========")
		assert.Contains(t, line, `"level":"trace"`)
		assert.Contains(t, line, "zlog.TestTRACESkip")
		assert.Regexp(t, `zlog/trace_test\.go:[89]\d`, line)
	}
}

func wrappedTRACE(args ...any) {

	var wrappedLevel2 = func(args ...any) {
		TRACESkip(2, args...)
		return
	}

	TRACESkip(1, args...)
	wrappedLevel2(args...)
}

func wrappedTRACE1(arg0 any, args ...any) {

	var wrappedLevel2 = func(arg0 any, args ...any) {
		TRACESkip1(2, arg0, args...)
		return
	}

	TRACESkip1(1, arg0, args...)
	wrappedLevel2(arg0, args...)
}

func TestTRACE1(t *testing.T) {
	buf := &zaptest.Buffer{}
	l, p, _ := NewWithOutput(&Config{Level: "trace", Development: true}, buf)
	defer ReplaceGlobals(l, p)()

	TRACE1("a")
	got := buf.String()
	assert.Contains(t, got, "zlog/trace_test.go:131")

	buf.Reset()
	TRACE1(WithLogger(context.Background(), L().With(zap.Int("TestTRACE1", 1))))
	TRACE1(L().With(zap.Int("TestTRACE1", 1)))
	TRACE1(S().With("TestTRACE1", 1))
	for _, line := range buf.Lines() {
		assert.Contains(t, line, " TRACE ")
		assert.Regexp(t, `zlog/trace_test.go:13\d`, line)
		assert.Contains(t, line, `"TestTRACE1": 1`)
	}

	buf.Reset()
	TRACE1("a", "b", "c", 1, 2, 3)
	TRACE1(context.Background(), "a", "b", "c", 1, 2, 3)
	TRACE1(L(), "a", "b", "c", 1, 2, 3)
	TRACE1(S(), "a", "b", "c", 1, 2, 3)
	for _, line := range buf.Lines() {
		assert.Contains(t, line, " TRACE ")
		assert.Regexp(t, `zlog/trace_test.go:14\d`, line)
		assert.Contains(t, line, "abc1 2 3")
	}

	buf.Reset()
	TRACE1("a=%v, b=%v, c=%v", 1, 2, 3)
	TRACE1(context.Background(), "a=%v, b=%v, c=%v", 1, 2, 3)
	TRACE1(L(), "a=%v, b=%v, c=%v", 1, 2, 3)
	TRACE1(S(), "a=%v, b=%v, c=%v", 1, 2, 3)
	for _, line := range buf.Lines() {
		assert.Contains(t, line, " TRACE ")
		assert.Regexp(t, `zlog/trace_test.go:1[56]\d`, line)
		assert.Contains(t, line, "a=1, b=2, c=3")
	}
}

func TestTRACESkip1(t *testing.T) {
	buf := &zaptest.Buffer{}
	l, p, _ := NewWithOutput(&Config{Level: "trace"}, buf)
	defer ReplaceGlobals(l, p)()

	TRACE1("test trace 1")
	TRACESkip1(0, "test trace 1")
	wrappedTRACE1("test trace 1") // this line outputs two messages

	lines := buf.Lines()
	assert.Len(t, lines, 4)
	for _, line := range lines {
		t.Log(line)
		assert.Contains(t, line, `"level":"trace"`)
		assert.Contains(t, line, "test trace 1")
		assert.Contains(t, line, `"level":"trace"`)
		assert.Contains(t, line, "zlog.TestTRACESkip1")
		assert.Regexp(t, `zlog/trace_test\.go:17\d`, line)
	}
}

func TestTraceFilterRule(t *testing.T) {
	msg := "test TraceFilterRule"
	testCases := []struct {
		name     string
		rule     string
		contains bool
	}{
		{name: "default", rule: "", contains: true},
		{name: "allow all", rule: "allow=all", contains: true},
		{name: "allow explicitly 1", rule: "allow=zlog/*.go", contains: true},
		{name: "allow explicitly 2", rule: "allow=gopkg/**", contains: true},
		{name: "not allowed explicitly", rule: "allow=confr/*,easy/*.go,easy/ezhttp/**", contains: false},
		{name: "deny all", rule: "deny=all", contains: false},
		{name: "deny explicitly", rule: "deny=zlog/*", contains: false},
		{name: "not denied explicitly", rule: "deny=confr/*,easy/ezhttp/**", contains: true},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &zaptest.Buffer{}
			cfg := &Config{Level: "trace", Format: "logfmt"}
			cfg.TraceFilterRule = tc.rule
			l, p, _ := NewWithOutput(cfg, buf)
			defer ReplaceGlobals(l, p)()

			L().Trace(msg)
			S().Tracef(msg)
			TRACE(msg)
			got := buf.String()
			if tc.contains {
				assert.Contains(t, got, msg)
				assert.Equal(t, 3, strings.Count(got, msg))
			} else {
				assert.NotContains(t, got, msg)
			}
		})
	}
}
