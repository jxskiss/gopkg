package easy

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"strings"
	"sync"
	"testing"
)

var _ = log.Println

var wantPanicLoc string

func willPanic() {
	wantPanicLoc = "gopkg/easy.willPanic:20"
	panic("oops...")
}

func willPanicCaller() {
	willPanic()
}

func TestIdentifyPanicLo(t *testing.T) {
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

type bufLogger struct {
	bytes.Buffer
	wg sync.WaitGroup
}

func (p *bufLogger) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(&p.Buffer, format, args...)
	p.wg.Done()
}

func TestGoroutineRecover(t *testing.T) {
	var logger = &bufLogger{}
	logger.wg.Add(1)
	Go(func() {
		willPanic()
	}, logger)
	logger.wg.Wait()
	logText := logger.String()
	assert.Contains(t, logText, "catch panic:")
	assert.Contains(t, logText, "gopkg/easy.TestGoroutineRecover")

	logger.Reset()
	logger.wg.Add(1)
	Go1(func() error {
		willPanicCaller()
		return nil
	}, logger)
	logger.wg.Wait()
	logText = logger.String()
	assert.Contains(t, logText, "catch panic:")
	assert.Contains(t, logText, "gopkg/easy.TestGoroutineRecover")

	logger.Reset()
	logger.wg.Add(1)
	Go1(func() error {
		return errors.New("dummy error")
	}, logger)
	logger.wg.Wait()
	logText = logger.String()
	assert.Contains(t, logText, "catch error:")
	assert.Contains(t, logText, "dummy error")
}
