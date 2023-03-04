package zlog

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newTestLogrLogger(lv Level, w io.Writer) *zap.Logger {
	if w == nil {
		w = io.Discard
	}
	cfg := &Config{
		Level:             lv.String(),
		Format:            "json",
		DisableTimestamp:  true,
		DisableStacktrace: true,
	}
	logger, _, err := NewWithOutput(cfg, zapcore.AddSync(w))
	if err != nil {
		panic(err)
	}
	return logger
}

func TestR(t *testing.T) {
	r0 := R()
	assert.NotNil(t, r0.GetSink().(*logrImpl).c)
	assert.NotNil(t, r0.GetSink().(*logrImpl).l)

	cfg := &LogrConfig{
		ErrorKey: "err",
	}
	r1 := R(cfg)
	assert.Equal(t, cfg, r1.GetSink().(*logrImpl).c)
	assert.NotNil(t, r1.GetSink().(*logrImpl).l)
	r11 := R(*cfg)
	assert.Equal(t, "err", r11.GetSink().(*logrImpl).c.ErrorKey)
	assert.NotNil(t, r1.GetSink().(*logrImpl).l)

	l := L().With(zap.String("ns", "default"))
	s := l.Sugar()
	r2 := R(l)
	r2.Info("test R(*zap.Logger)")
	r3 := R(s)
	r3.Info("test R(*zap.SugaredLogger)")

	b := B().With(zap.Int("podnum", 2))
	ctx := WithBuilder(context.Background(), b)
	r4 := R(b)
	r4.Info("test R(*Builder)")
	r5 := R(ctx)
	r5.Info("test R(context.Context)")
}

func TestLogrLoggerInfo(t *testing.T) {
	var buffer bytes.Buffer
	writer := bufio.NewWriter(&buffer)
	zl := newTestLogrLogger(TraceLevel, writer)
	testLogger := R(&LogrConfig{
		DPanicOnInvalidLog: true,
		Logger:             zl,
	})

	testLogger.Info("test info", "ns", "default", "podnum", 2)
	err := writer.Flush()
	require.Nil(t, err)

	logStr := buffer.String()
	assert.Contains(t, logStr, `"level":"info"`)
	assert.Contains(t, logStr, `"caller":"zlog/logr_test.go:`)
	assert.Contains(t, logStr, `"msg":"test info"`)
	assert.Contains(t, logStr, `"ns":"default"`)
	assert.Contains(t, logStr, `"podnum":2`)

	// test invalid log
	buffer.Reset()
	testLogger.Info("invalid", zap.String("ns", "default"), 12345)
	err = writer.Flush()
	require.Nil(t, err)

	logStr = buffer.String()
	fmt.Println(logStr)
	assert.Contains(t, logStr, `"level":"dpanic"`)
	assert.Contains(t, logStr, `"caller":"zlog/logr_test.go:`)
	assert.Contains(t, logStr, `"msg":"invalid"`)
	assert.Contains(t, logStr, `"ns":"default"`)
	assert.Contains(t, logStr, `"ignoredKey":12345`)

	buffer.Reset()
	testLogger.Info("invalid", zap.String("ns", "default"), 12345, "abcde")
	err = writer.Flush()
	require.Nil(t, err)

	logStr = buffer.String()
	fmt.Println(logStr)
	assert.Contains(t, logStr, `"level":"dpanic"`)
	assert.Contains(t, logStr, `"caller":"zlog/logr_test.go:`)
	assert.Contains(t, logStr, `"msg":"invalid"`)
	assert.Contains(t, logStr, `"ns":"default"`)
	assert.Contains(t, logStr, `"invalidKey":12345`)
}

func TestLogrLoggerError(t *testing.T) {
	for _, logErrKey := range []string{
		"err",
		"error",
	} {
		t.Run(fmt.Sprintf("error field name %s", logErrKey), func(t *testing.T) {
			var buffer bytes.Buffer
			writer := bufio.NewWriter(&buffer)
			zl := newTestLogrLogger(InfoLevel, writer)
			testLogger := R(&LogrConfig{
				ErrorKey:        logErrKey,
				NumericLevelKey: "v",
				Logger:          zl,
			})

			// Errors always get logged, regardless of log levels.
			testLogger.V(10).Error(fmt.Errorf("invalid namespace:%s", "default"), "wrong namespace", "ns", "default", "podnum", 2)
			err := writer.Flush()
			require.Nil(t, err)

			logStr := buffer.String()
			assert.Contains(t, logStr, `"level":"error"`)
			assert.Contains(t, logStr, `"caller":"zlog/logr_test.go:`)
			assert.Contains(t, logStr, `"msg":"wrong namespace"`)
			assert.Contains(t, logStr, `"ns":"default"`)
			assert.Contains(t, logStr, `"podnum":2`)
		})
	}
}

func TestLogrLoggerEnabled(t *testing.T) {
	for i := 0; i < 11; i++ {
		t.Run(fmt.Sprintf("logger level %d", i), func(t *testing.T) {
			cfg := &LogrConfig{
				Logger: newTestLogrLogger(Level(-i), nil),
			}
			testLogger := R(cfg)

			for j := 0; j <= 128; j++ {
				shouldBeEnabled := i >= j
				t.Run(fmt.Sprintf("message level %d", j), func(t *testing.T) {
					isEnabled := testLogger.V(j).Enabled()
					if !isEnabled && shouldBeEnabled {
						t.Errorf("V(%d).Info should be enabled", j)
					} else if isEnabled && !shouldBeEnabled {
						t.Errorf("V(%d).Info should not be enabled", j)
					}

					log := testLogger
					for k := 0; k < j; k++ {
						log = log.V(1)
					}
					isEnabled = log.Enabled()
					if !isEnabled && shouldBeEnabled {
						t.Errorf("repeated V(1).Info should be enabled")
					} else if isEnabled && !shouldBeEnabled {
						t.Errorf("repeated V(1).Info should not be enabled")
					}
				})
			}
		})
	}
}

func TestLogrNumericLevel(t *testing.T) {
	for _, logNumKey := range []string{
		"",
		"v",
		"verbose",
	} {
		t.Run(fmt.Sprintf("numeric verbosity field %q", logNumKey), func(t *testing.T) {
			for i := 0; i < 4; i++ {
				var buffer bytes.Buffer
				writer := bufio.NewWriter(&buffer)
				cfg := &LogrConfig{
					NumericLevelKey: logNumKey,
					Logger:          newTestLogrLogger(Level(-100), writer),
				}
				testLogger := R(cfg)

				testLogger.V(i).Info("test", "ns", "default", "podnum", 2)
				err := writer.Flush()
				require.Nil(t, err)

				logStr := buffer.String()
				fmt.Println(logStr)

				assert.Contains(t, logStr, `"caller":"zlog/logr_test.go:`)
				if logNumKey != "" {
					assert.Contains(t, logStr, fmt.Sprintf(`"%s":%d`, logNumKey, i))
				}
			}
		})
	}
}
