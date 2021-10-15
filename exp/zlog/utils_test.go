package zlog

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestTRACE(t *testing.T) {
	defer replaceGlobals(mustNewGlobalLogger(&Config{Level: "trace"}))()

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

func TestDEBUG(t *testing.T) {
	defer replaceGlobals(mustNewGlobalLogger(&Config{Level: "trace"}))()

	DEBUG()
	DEBUG(context.Background())
	DEBUG(L())
	DEBUG(S())
}

func TestTRACESkip(t *testing.T) {
	buf := &zaptest.Buffer{}
	l, p, _ := NewWithOutput(&Config{Level: "trace"}, buf)
	defer replaceGlobals(l, p)()

	TRACE()
	TRACESkip(0)
	wrappedTRACE() // this line outputs two messages

	lines := buf.Lines()
	assert.Len(t, lines, 4)
	for _, line := range lines {
		assert.Contains(t, line, "========")
		assert.Contains(t, line, "zlog/utils_test.go:")
		assert.Contains(t, line, "zlog.TestTRACESkip")
	}
}

func wrappedTRACE(args ...interface{}) {

	var wrappedLevel2 = func(args ...interface{}) {
		TRACESkip(2, args...)
		return
	}

	TRACESkip(1, args...)
	wrappedLevel2(args...)
}
