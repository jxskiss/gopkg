package zlog

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func TestTrace(t *testing.T) {
	buf := &zaptest.Buffer{}
	l, p, _ := NewWithOutput(&Config{Level: "trace", Format: "logfmt"}, buf)
	defer ReplaceGlobals(l, p)()

	Trace("trace message", zap.String("k1", "v1"))

	got := buf.String()
	assert.Contains(t, got, "level=trace")
	assert.Contains(t, got, "k1=v1")
	assert.Contains(t, got, "caller=zlog/trace_test.go:17")
	assert.Contains(t, got, `msg="[TRACE] trace message"`)
}

func TestTracef(t *testing.T) {
	buf := &zaptest.Buffer{}
	l, p, _ := NewWithOutput(&Config{Level: "trace", Format: "logfmt"}, buf)
	defer ReplaceGlobals(l, p)()

	Tracef("trace message, %v, %v", 123, 456)

	got := buf.String()
	assert.Contains(t, got, "level=trace")
	assert.Contains(t, got, "caller=zlog/trace_test.go:31")
	assert.Contains(t, got, `msg="[TRACE] trace message, 123, 456"`)
}

func TestTRACE(t *testing.T) {
	defer ReplaceGlobals(mustNewGlobalLogger(&Config{Level: "trace", Development: true}))()

	TRACE()
	TRACE(context.Background())
	TRACE(L())
	TRACE(S())

	TRACE("a", "b", "c", 1, 2, 3)
	TRACE(context.Background(), "a", "b", "c", 1, 2, 3)
	TRACE(L(), "a", "b", "c", 1, 2, 3)
	TRACE(S(), "a", "b", "c", 1, 2, 3)

	TRACE("a=%v, b=%v, c=%v", 1, 2, 3)
	TRACE(context.Background(), "a=%v, b=%v, c=%v", 1, 2, 3)
	TRACE(L(), "a=%v, b=%v, c=%v", 1, 2, 3)
	TRACE(S(), "a=%v, b=%v, c=%v", 1, 2, 3)
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
		t.Log(line)
		assert.Contains(t, line, `"level":"trace"`)
		assert.Contains(t, line, "========")
		assert.Contains(t, line, TracePrefix)
		assert.Contains(t, line, "zlog.TestTRACESkip")
		assert.Regexp(t, `zlog/trace_test\.go:6[3-5]`, line)
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
	defer ReplaceGlobals(mustNewGlobalLogger(&Config{Level: "trace", Development: true}))()

	TRACE1(context.Background())
	TRACE1(L())
	TRACE1(S())

	TRACE1("a", "b", "c", 1, 2, 3)
	TRACE1(context.Background(), "a", "b", "c", 1, 2, 3)
	TRACE1(L(), "a", "b", "c", 1, 2, 3)
	TRACE1(S(), "a", "b", "c", 1, 2, 3)

	TRACE1("a=%v, b=%v, c=%v", 1, 2, 3)
	TRACE1(context.Background(), "a=%v, b=%v, c=%v", 1, 2, 3)
	TRACE1(L(), "a=%v, b=%v, c=%v", 1, 2, 3)
	TRACE1(S(), "a=%v, b=%v, c=%v", 1, 2, 3)
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
		assert.Contains(t, line, TracePrefix)
		assert.Contains(t, line, "zlog.TestTRACESkip1")
		assert.Regexp(t, `zlog/trace_test\.go:12[4-6]`, line)
	}
}
