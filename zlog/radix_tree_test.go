package zlog

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestRadixTree(t *testing.T) {
	root := &radixNode[Level]{}
	root.insert("some.module_2.pkg_1", InfoLevel)
	root.insert("some.module_2.pkg_2", DebugLevel)
	root.insert("zlog.filtertest", WarnLevel)
	root.insert("some.module_1", ErrorLevel)
	root.insert("some.module_1.pkg_1", WarnLevel)

	assert.Equal(t, `some.module_1=error
some.module_1.pkg_1=warn
some.module_2.pkg_1=info
some.module_2.pkg_2=debug
zlog.filtertest=warn`, strings.TrimSpace(root.dumpTree("")))

	testcases := []struct {
		Name  string
		Level Level
		Found bool
	}{
		{"some.module_1", ErrorLevel, true},
		{"some.module_1.pkg_0", ErrorLevel, true},
		{"some.module_1.pkg_1", WarnLevel, true},
		{"some.module_2", InfoLevel, false},
		{"some.module_2.pkg_0", InfoLevel, false},
		{"some.module_2.pkg_1", InfoLevel, true},
		{"some.module_2.pkg_2", DebugLevel, true},
	}
	for _, tc := range testcases {
		level, found := root.search(tc.Name)
		assert.Equalf(t, tc.Level, level, "name= %v", tc.Name)
		assert.Equal(t, tc.Found, found, "name= %v", tc.Name)
	}
}

func TestPerLoggerLevels(t *testing.T) {
	buf := &zaptest.Buffer{}
	filters := []string{
		"zlog.testfilters=trace",
		"some.pkg2=error",
	}
	logger, _, err := NewWithOutput(&Config{Development: true, Level: "info", PerLoggerLevels: filters}, buf)
	assert.Nil(t, err)
	SetLevel(TraceLevel)

	lg1 := logger.Named("zlog")
	lg1.Debug("zlog debug message 1") // no
	lg1.Info("zlog info message 2")   // yes

	lg1 = lg1.Named("testfilters")
	lg1.Debug("zlog debug message 3") // yes
	lg1.Info("zlog info message 4")   // yes
	TRACE(lg1, "zlog TRACE message")  // yes

	lg2 := logger.Named("some.pkg2")
	lg2.Info("zlog info message 5")   // no
	lg2.Warn("zlog warn message 6")   // no
	lg2.Error("zlog error message 7") // yes

	got := buf.String()
	assert.NotContains(t, got, "zlog debug message 1")
	assert.Contains(t, got, "zlog info message 2")
	assert.Contains(t, got, "zlog debug message 3")
	assert.Contains(t, got, "zlog info message 4")
	assert.Contains(t, got, "zlog TRACE message")
	assert.NotContains(t, got, "zlog info message 5")
	assert.NotContains(t, got, "zlog warn message 6")
	assert.Contains(t, got, "zlog error message 7")
}
