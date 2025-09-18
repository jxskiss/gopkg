package easy

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteFileAndReadFileLines(t *testing.T) {
	lines := `
line1
line two
`
	tmpDir := t.TempDir()
	filename := filepath.Join(tmpDir, "a/b", "ccc.txt")
	err := WriteFile(filename, []byte(lines), 0o644)
	require.Nil(t, err)

	gotLines, err := ReadFileLines(filename)
	require.Nil(t, err)
	require.Len(t, gotLines, 3)
	assert.Equal(t, "", gotLines[0])
	assert.Equal(t, "line1", gotLines[1])
	assert.Equal(t, "line two", gotLines[2])
}

func Test_getDirectoryPermFromFilePerm(t *testing.T) {
	testData := []struct {
		FilePerm os.FileMode
		Want     os.FileMode
	}{
		{0o020, 0o730},
		{0o040, 0o750},
		{0o060, 0o770},
		{0o002, 0o703},
		{0o004, 0o705},
		{0o006, 0o707},

		{0o644, 0o755},
		{0o666, 0o777},
	}
	for _, tCase := range testData {
		got := getDirectoryPermFromFilePerm(tCase.FilePerm)
		assert.Equal(t, tCase.Want, got)
	}
}

func TestRunTask(t *testing.T) {
	t.Run("task finished", func(t *testing.T) {
		var taskDone bool
		var signalReceived bool
		task := func() {
			time.Sleep(time.Second)
			taskDone = true
		}
		onSignal := func(_ os.Signal) {
			signalReceived = true
		}
		RunTask(task, onSignal)
		assert.True(t, taskDone)
		assert.False(t, signalReceived)
	})
}
