package timeutil

import (
	"bytes"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLatencyRecorder(t *testing.T) {
	lRec := NewLatencyRecorder()
	time.Sleep(10 * time.Millisecond)
	lRec.Mark("op1")
	time.Sleep(10 * time.Millisecond)
	lRec.Mark("op2")
	lRec.MarkFromStartTime("op3")
	lRec.MarkWithStartTime("op4", time.Now().Add(-time.Second))
	time.Sleep(time.Millisecond)

	var buf bytes.Buffer
	marks, latencyMap := lRec.GetLatencyMap()
	logStr := lRec.Format()
	_, err := lRec.WriteTo(&buf)
	t.Logf("marks: %v", marks)
	t.Logf("logStr: %s", logStr)
	t.Logf("buf: %s", buf.String())

	require.Nil(t, err)
	assert.Equal(t, []string{"op1", "op2", "op3", "op4", "total"}, marks)
	assert.GreaterOrEqual(t, latencyMap["op1"], 10*time.Millisecond)
	assert.GreaterOrEqual(t, latencyMap["op2"], 10*time.Millisecond)
	// Timer on Windows platform is inaccurate, skip this assertion,
	// it fails randomly on Windows platform.
	if runtime.GOOS != "windows" {
		assert.Less(t, latencyMap["op2"], 20*time.Millisecond)
	}
	assert.GreaterOrEqual(t, latencyMap["op3"], 20*time.Millisecond)
	assert.GreaterOrEqual(t, latencyMap["op4"], time.Second)
	assert.Greater(t, latencyMap["total"], 20*time.Millisecond)
	assert.Regexp(t, `op1=\d+`, logStr)
	assert.Regexp(t, `\s+op2=\d+`, logStr)
	assert.Regexp(t, `\s+op3=\d+`, logStr)
	assert.Regexp(t, `\s+op4=\d{4}`, logStr)
	assert.Regexp(t, `\s+total=\d+`, logStr)

	lRec.Reset()
	_, latencyMap = lRec.GetLatencyMap()
	assert.Len(t, latencyMap, 1)
	assert.Contains(t, latencyMap, "total")
}
