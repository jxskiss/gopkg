package zlog

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jxskiss/gopkg/v2/utils/ptr"
)

func TestMultiFilesCore(t *testing.T) {

	getTempFilename := func() string {
		tmpFilenamePattern := "zlog-test-multi-files-core-*.log"
		f1, err := os.CreateTemp("", tmpFilenamePattern)
		require.Nil(t, err)
		f1Name := f1.Name()
		f1.Close()
		return f1Name
	}

	removeFiles := func(filenames ...string) {
		for _, fn := range filenames {
			os.Remove(fn)
		}
	}

	f1 := getTempFilename()
	f2 := getTempFilename()
	f3 := getTempFilename()

	removeFiles(f1, f2, f3)
	defer removeFiles(f1, f2, f3)

	cfg := &Config{
		File: FileLogConfig{
			Filename:   f1,
			MaxSize:    ptr.Int(100),
			MaxDays:    ptr.Int(3),
			MaxBackups: nil,
			LocalTime:  ptr.Bool(true),
			Compress:   nil,
		},
		PerLoggerFiles: map[string]FileLogConfig{
			"access": {Filename: f2},
			"pkg2":   {Filename: f3},
		},
	}

	logger, _, err := New(cfg)
	require.Nil(t, err)

	logger.Info("from default logger")

	lg1 := logger.Named("access")
	lg1.Info("from access logger")
	lg1.Named("sub.logger1").Info("from access.sub.logger1 logger")
	logger.Named("access.sub.logger2").Info("from access.sub.logger2 logger")

	lg2 := logger.Named("pkg2")
	lg2.Info("from pkg2 logger")
	lg2.Named("sub").Info("from pkg2.sub logger")

	lg3 := logger.Named("pkg3")
	lg3.Info("from pkg3 logger")
	lg3.Named("sub").Info("from pkg3.sub logger")

	readFile := func(filename string) string {
		data, err := os.ReadFile(filename)
		require.Nil(t, err)
		return string(data)
	}

	got1 := readFile(f1)
	assert.Contains(t, got1, "from default logger")
	assert.Contains(t, got1, "from pkg3 logger")
	assert.Contains(t, got1, "from pkg3.sub logger")
	assert.NotContains(t, got1, "from access")
	assert.NotContains(t, got1, "from pkg2")

	got2 := readFile(f2)
	assert.NotContains(t, got2, "from default")
	assert.NotContains(t, got2, "from pkg2")
	assert.NotContains(t, got2, "from pkg3")
	assert.Contains(t, got2, "from access logger")
	assert.Contains(t, got2, "from access.sub.logger1 logger")
	assert.Contains(t, got2, "from access.sub.logger2 logger")

	got3 := readFile(f3)
	assert.NotContains(t, got3, "from default")
	assert.NotContains(t, got3, "from access")
	assert.NotContains(t, got3, "from pkg3")
	assert.Contains(t, got3, "from pkg2 logger")
	assert.Contains(t, got3, "from pkg2.sub logger")
}
