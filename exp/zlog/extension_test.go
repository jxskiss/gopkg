package zlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

func BenchmarkLTracef_AddCallerSkip(b *testing.B) {
	logger := L().Sugar()

	for i := 0; i < b.N; i++ {
		_ = logger.Desugar().WithOptions(zap.AddCallerSkip(1)).Sugar()
	}
}

func TestTrace_AddCallerSkip(t *testing.T) {
	buf := &zaptest.Buffer{}
	cfg := &Config{Level: "trace", Format: "console", DisableTimestamp: true}
	l, p, err := NewWithOutput(cfg, buf)
	if err != nil {
		panic(err)
	}
	defer replaceGlobals(l, p)()

	msg := "TestTrace_AddCallerSkip msg"
	Trace(msg)
	Tracef(msg)
	LTrace(L(), msg)
	LTracef(S(), msg)

	outputLines := buf.Lines()
	assert.Len(t, outputLines, 4)
	for _, line := range outputLines {
		assert.Contains(t, line, "[Trace] "+msg)
		assert.Contains(t, line, "zlog/extension_test.go")
	}
}
