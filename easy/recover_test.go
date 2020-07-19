package easy

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"log"
	"strings"
	"sync"
	"testing"
)

var _ = log.Println

var wantPanicLoc string

func willPanic() {
	wantPanicLoc = "gopkg/easy.willPanic:18"
	panic("oops...")
}

func willPanicCaller() {
	willPanic()
}

func TestIdentifyPanicLoc(t *testing.T) {
	var panicLoc1 string
	func() {
		defer func() {
			recover()
			panicLoc1 = IdentifyPanic()
		}()
		willPanic()
	}()
	assert.True(t, strings.HasSuffix(panicLoc1, wantPanicLoc))

	var panicLoc2 string
	func() {
		defer func() {
			recover()
			panicLoc2 = IdentifyPanic()
		}()
		willPanicCaller()
	}()
	assert.True(t, strings.HasSuffix(panicLoc2, wantPanicLoc))
}

type syncLogger struct {
	bufLogger
	wg sync.WaitGroup
}

func (p *syncLogger) Errorf(format string, args ...interface{}) {
	p.bufLogger.Errorf(format, args...)
	p.wg.Done()
}

func TestGoroutineRecover(t *testing.T) {
	var logger = &syncLogger{}
	ConfigLog(false, logger, nil)

	logger.wg.Add(1)
	Go(func() { willPanic() })
	logger.wg.Wait()
	logText := logger.buf.String()
	assert.Contains(t, logText, "catch panic:")
	assert.Contains(t, logText, "gopkg/easy.TestGoroutineRecover")

	logger.buf.Reset()
	logger.wg.Add(1)
	Go1(func() error {
		willPanicCaller()
		return nil
	})
	logger.wg.Wait()
	logText = logger.buf.String()
	assert.Contains(t, logText, "catch panic:")
	assert.Contains(t, logText, "gopkg/easy.TestGoroutineRecover")

	logger.buf.Reset()
	logger.wg.Add(1)
	Go1(func() error {
		return errors.New("dummy error")
	})
	logger.wg.Wait()
	logText = logger.buf.String()
	assert.Contains(t, logText, "catch error:")
	assert.Contains(t, logText, "dummy error")
}

func TestPanicOnError(t *testing.T) {
	panicErr := errors.New("dummy panic error")
	willPanic := func() (int, error) {
		return 123, panicErr
	}

	x, gotErr := willPanic()
	assert.PanicsWithValue(t, panicErr, func() {
		PanicOnError(x, gotErr)
	})

	assert.PanicsWithValue(t, panicErr, func() {
		Must(willPanic())
	})
}
