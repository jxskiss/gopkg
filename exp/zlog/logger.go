package zlog

import (
	"fmt"
	"log"
	"os"
	_ "unsafe"

	"go.uber.org/zap"
)

var _ Logger = (*zap.SugaredLogger)(nil)

// Logger is a generic logger interface that output logs with a format.
// It's implemented by many logging libraries, including logrus.Logger,
// zap.SugaredLogger, etc.
//
// Within this package, StdLogger is a default implementation which sends
// log messages to the standard library.
type Logger interface {
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
}

// -------- standard library logger -------- //

// StdLogger is a default implementation of Logger which sends log messages
// to the standard library.
var StdLogger Logger = &stdLogger{}

type stdLogger struct{}

const _stdLogDepth = 2

// log_std links to log.std to get correct caller depth for both
// with and without calling RedirectStdLog.
//go:linkname log_std log.std
var log_std *log.Logger

func (s stdLogger) Debugf(format string, args ...interface{}) {
	if GetLevel() <= DebugLevel {
		log_std.Output(_stdLogDepth, fmt.Sprintf("[Debug] "+format, args...))
	}
}

func (s stdLogger) Infof(format string, args ...interface{}) {
	if GetLevel() <= InfoLevel {
		log_std.Output(_stdLogDepth, fmt.Sprintf("[Info] "+format, args...))
	}
}

func (s stdLogger) Warnf(format string, args ...interface{}) {
	if GetLevel() <= WarnLevel {
		log_std.Output(_stdLogDepth, fmt.Sprintf("[Warn] "+format, args...))
	}
}

func (s stdLogger) Errorf(format string, args ...interface{}) {
	if GetLevel() <= ErrorLevel {
		log_std.Output(_stdLogDepth, fmt.Sprintf("[Error] "+format, args...))
	}
}

func (s stdLogger) Fatalf(format string, args ...interface{}) {
	log_std.Output(_stdLogDepth, fmt.Sprintf("[Fatal] "+format, args...))
	os.Exit(1)
}
