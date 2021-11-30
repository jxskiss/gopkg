package zlog

import (
	"fmt"
	"os"

	"go.uber.org/zap"

	"github.com/jxskiss/gopkg/v2/internal/linkname"
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
var StdLogger Logger = stdLogger{}

type stdLogger struct{}

// log_std links to log.std to get correct caller depth for both
// with and without calling RedirectStdLog.
var log_std = linkname.LogStd

const _stdLogDepth = 2

func (_ stdLogger) Debugf(format string, args ...interface{}) {
	if GetLevel() <= DebugLevel {
		log_std.Output(_stdLogDepth, fmt.Sprintf(DebugPrefix+format, args...))
	}
}

func (_ stdLogger) Infof(format string, args ...interface{}) {
	if GetLevel() <= InfoLevel {
		log_std.Output(_stdLogDepth, fmt.Sprintf(InfoPrefix+format, args...))
	}
}

func (_ stdLogger) Warnf(format string, args ...interface{}) {
	if GetLevel() <= WarnLevel {
		log_std.Output(_stdLogDepth, fmt.Sprintf(WarnPrefix+format, args...))
	}
}

func (_ stdLogger) Errorf(format string, args ...interface{}) {
	if GetLevel() <= ErrorLevel {
		log_std.Output(_stdLogDepth, fmt.Sprintf(ErrorPrefix+format, args...))
	}
}

func (_ stdLogger) Fatalf(format string, args ...interface{}) {
	log_std.Output(_stdLogDepth, fmt.Sprintf(FatalPrefix+format, args...))
	Sync()
	os.Exit(1)
}

// -------- nop logger -------- //

// NopLogger is a logger which discards anything it receives.
var NopLogger Logger = &nopLogger{}

type nopLogger struct{}

func (_ nopLogger) Debugf(format string, args ...interface{}) {}

func (_ nopLogger) Infof(format string, args ...interface{}) {}

func (_ nopLogger) Warnf(format string, args ...interface{}) {}

func (_ nopLogger) Errorf(format string, args ...interface{}) {}

func (_ nopLogger) Fatalf(format string, args ...interface{}) {
	Sync()
	os.Exit(1)
}
