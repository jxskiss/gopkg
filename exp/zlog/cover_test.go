package zlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func Test_Cover_globalLoggingFunctions(t *testing.T) {
	buf := &zaptest.Buffer{}
	l, p, err := NewWithOutput(&Config{Level: "trace"}, buf)
	if err != nil {
		panic(err)
	}
	defer replaceGlobals(l, p)()

	msg := "cover message"
	Trace(msg)
	Tracef(msg)
	Debug(msg)
	Debugf(msg)
	Info(msg)
	Infof(msg)
	Warn(msg)
	Warnf(msg)
	Error(msg)
	Errorf(msg)

	outputLines := buf.Lines()
	assert.Len(t, outputLines, 10)
	for _, line := range outputLines {
		assert.Contains(t, line, msg)
		assert.Contains(t, line, "zlog/cover_test.go:")
	}
}
