package zlog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func Test_Cover_globalLoggingFunctions(t *testing.T) {
	buf := &zaptest.Buffer{}
	l, p, err := NewWithOutput(&Config{Level: "trace"}, buf)
	require.Nil(t, err)
	defer ReplaceGlobals(l, p)()

	msg := "cover message"
	Debug(msg)
	Debugf(msg)
	Info(msg)
	Infof(msg)
	Warn(msg)
	Warnf(msg)
	Error(msg)
	Errorf(msg)
	Print(msg)
	Printf(msg)
	Println(msg)

	err = Sync()
	require.Nil(t, err)

	outputLines := buf.Lines()
	assert.Len(t, outputLines, 11)
	for _, line := range outputLines {
		assert.Contains(t, line, msg)
		assert.Contains(t, line, "zlog/cover_test.go:")
	}
}
