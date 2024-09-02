package zlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestGlobalLoggingFunctions(t *testing.T) {
	buf := &zaptest.Buffer{}
	l, p, err := NewWithOutput(&Config{Level: "trace", Format: "logfmt"}, buf)
	require.Nil(t, err)
	defer ReplaceGlobals(l, p)()

	msg := "cover message"
	Debug(msg)
	Debugf(msg)
	Debugw(msg, "key1", "value1")
	Info(msg)
	Infof(msg)
	Infow(msg, "key1", "value1")
	Warn(msg)
	Warnf(msg)
	Warnw(msg, "key1", "value1")
	Error(msg)
	Errorf(msg)
	Errorw(msg, "key1", "value1")
	Print(msg)
	Printf(msg)
	Println(msg)

	err = Sync()
	require.Nil(t, err)

	outputLines := buf.Lines()
	assert.Len(t, outputLines, 15)
	for _, line := range outputLines {
		assert.Contains(t, line, msg)
		assert.Contains(t, line, "zlog/global_test.go:")
	}
}
