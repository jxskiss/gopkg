//nolint:errcheck
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
// log messages to the standard library, it also adds the level prefix to
// the output message.
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
//
// It follows the global logging level of this package, the level can be
// changed by calling SetLevel.
var StdLogger Logger = stdLogger{}

type stdLogger struct{}

// log_std links to log.std to get correct caller depth for both
// with and without setting GlobalConfig.RedirectStdLog.
//
//go:linkname log_std log.std
var log_std *log.Logger

const _stdLogDepth = 2

func (stdLogger) Debugf(format string, args ...interface{}) {
	if GetLevel() <= DebugLevel {
		log_std.Output(_stdLogDepth, fmt.Sprintf(DebugPrefix+format, args...))
	}
}

func (stdLogger) Infof(format string, args ...interface{}) {
	if GetLevel() <= InfoLevel {
		log_std.Output(_stdLogDepth, fmt.Sprintf(InfoPrefix+format, args...))
	}
}

func (stdLogger) Warnf(format string, args ...interface{}) {
	if GetLevel() <= WarnLevel {
		log_std.Output(_stdLogDepth, fmt.Sprintf(WarnPrefix+format, args...))
	}
}

func (stdLogger) Errorf(format string, args ...interface{}) {
	if GetLevel() <= ErrorLevel {
		log_std.Output(_stdLogDepth, fmt.Sprintf(ErrorPrefix+format, args...))
	}
}

func (stdLogger) Fatalf(format string, args ...interface{}) {
	log_std.Output(_stdLogDepth, fmt.Sprintf(FatalPrefix+format, args...))
	Sync()
	os.Exit(1)
}

// -------- nop logger -------- //

// NopLogger is a logger which discards anything it receives.
var NopLogger Logger = &nopLogger{}

type nopLogger struct{}

func (nopLogger) Debugf(format string, args ...interface{}) {}

func (nopLogger) Infof(format string, args ...interface{}) {}

func (nopLogger) Warnf(format string, args ...interface{}) {}

func (nopLogger) Errorf(format string, args ...interface{}) {}

func (nopLogger) Fatalf(format string, args ...interface{}) {
	Sync()
	os.Exit(1)
}
