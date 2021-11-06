package zlog

import (
	"fmt"
	"strings"

	"go.uber.org/atomic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// A Level is a logging priority. Higher levels are more important.
type Level int8

const (
	// TraceLevel logs are the most fine-grained information which helps
	// developer "tracing" the code and trying to find one part of a
	// function specifically. Use this level when you need full visibility
	// of what is happening in your application and inside the third-party
	// libraries that you use.
	//
	// You can expect this logging level to be very verbose. You can use it
	// for example to annotate each step in the algorithm or each individual
	// query with parameters in your code.
	TraceLevel Level = iota

	// DebugLevel logs are less granular compared to TraceLevel, but it is
	// more than you will need in everyday use. DebugLevel should be used
	// for information that may be needed for diagnosing issues and
	// troubleshooting or when running application in development or test
	// environment for the purpose of making sure everything is running
	// correctly.
	//
	// DebugLevel logs are helpful for diagnosing and troubleshooting to
	// people more than just developers (e.g. IT, system admins, etc.).
	DebugLevel

	// InfoLevel logs are generally useful information which indicate that
	// something happened, the application entered a certain state, etc.
	// For example, a controller of your authorization API may write a
	// message at InfoLevel with information on which user requested
	// authorization if the authorization was successful or not.
	//
	// The information logged at InfoLevel should be purely informative
	// that you don't need to care about under normal circumstances, and
	// not looking into them on a regular basis shouldn't result in missing
	// any important information.
	//
	// This should be the out-of-box level for most applications in
	// service production deployment or application release configuration.
	InfoLevel

	// NoticeLevel logs are important information which should be always
	// available and shall not be turned off, user should be aware of these
	// events when they look into the system or application.
	// (Such as service start/stop/restart, reconnecting to database, switching
	// from a primary server to a backup server, retrying an operation, etc.)
	//
	// NoticeLevel is not implemented currently.
	NoticeLevel

	// WarnLevel logs indicate that something unexpected happened in the
	// system or application, a problem, or a situation that might
	// potentially cause application oddities, but the code can continue
	// to work. For example, unexpected disconnection from server, being
	// close to quota, suspicious web attach, temporarily heartbeat missing,
	// or a parsing error that resulted in a certain document not being
	// correctly processed.
	//
	// Warning messages may need human review, but generally that don't need
	// immediately intervention.
	WarnLevel

	// ErrorLevel logs indicate that the application hit issues preventing
	// one or more functionalities from properly functioning, they may be
	// fatal to an operation, but not fatal to the entire service or
	// application (e.g. can't open a required file, missing data,
	// temporarily failure from database or downstream service, etc.),
	// the application should continue running.
	//
	// These messages definitely need investigation and intervention from
	// user (developer, system administrator, or direct user), continuous
	// errors may cause serious problems, (e.g. service outage, lost of
	// income, or customer complaints, etc.).
	ErrorLevel

	// CriticalLevel logs indicate that the system or application encountered
	// critical condition preventing it to function properly, the system
	// or application is in a very unhealthy state.
	//
	// Intervention actions must be taken immediately, which means you should
	// go to get a system administrator or developer out of bed quickly.
	//
	// CriticalLevel is not implemented currently.
	CriticalLevel

	// DPanicLevel logs are particularly important errors.
	// In development mode the logger panics after writing the message.
	DPanicLevel

	// PanicLevel logs indicate that the application encountered unrecoverable
	// errors that it should abort immediately.
	//
	// The logger writes the message, then panics the application.
	PanicLevel

	// FatalLevel logs indicate that the application encountered unrecoverable
	// errors that it should abort immediately.
	//
	// The logger writes the message, then calls os.Exit to abort the application.
	FatalLevel
)

const (
	tracePrefix    = "[Trace] "
	debugPrefix    = "[Debug] "
	infoPrefix     = "[Info] "
	noticePrefix   = "[Notice] "
	warnPrefix     = "[Warn] "
	errorPrefix    = "[Error] "
	criticalPrefix = "[Critical] "
	fatalPrefix    = "[Fatal] "
)

const levelPrefixMinLen = 6

var mapZapLevels = [...]zapcore.Level{
	zap.DebugLevel,
	zap.DebugLevel,
	zap.InfoLevel,
	zap.InfoLevel,
	zap.WarnLevel,
	zap.ErrorLevel,
	zap.ErrorLevel,
	zap.DPanicLevel,
	zap.PanicLevel,
	zap.FatalLevel,
}

var levelNames = [...]string{
	"trace",
	"debug",
	"info",
	"notice",
	"warn",
	"error",
	"critical",
	"dpanic",
	"panic",
	"fatal",
}

func (l Level) toZapLevel() zapcore.Level { return mapZapLevels[l] }

func (l Level) String() string { return levelNames[l] }

func (l Level) Enabled(lvl zapcore.Level) bool {
	return lvl >= l.toZapLevel()
}

func (l *Level) unmarshalText(text []byte) bool {
	switch string(text) {
	case "trace", "TRACE":
		*l = TraceLevel
	case "debug", "DEBUG":
		*l = DebugLevel
	case "info", "INFO":
		*l = InfoLevel
	case "notice", "NOTICE":
		*l = NoticeLevel
	case "warn", "warning", "WARN", "WARNING":
		*l = WarnLevel
	case "error", "ERROR":
		*l = ErrorLevel
	case "critical", "CRITICAL":
		*l = CriticalLevel
	case "dpanic", "DPANIC":
		*l = DPanicLevel
	case "panic", "PANIC":
		*l = PanicLevel
	case "fatal", "FATAL":
		*l = FatalLevel
	default:
		return false
	}
	return true
}

type atomicLevel struct {
	lvl *atomic.Int32
	zl  zap.AtomicLevel
}

func newAtomicLevel() atomicLevel {
	return atomicLevel{
		lvl: atomic.NewInt32(int32(InfoLevel)),
		zl:  zap.NewAtomicLevel(),
	}
}

func (l atomicLevel) Level() Level { return Level(l.lvl.Load()) }

func (l *atomicLevel) SetLevel(lvl Level) {
	l.lvl.Store(int32(lvl))
	l.zl.SetLevel(lvl.toZapLevel())
}

func (l *atomicLevel) UnmarshalText(text []byte) error {
	var _lvl Level
	if !_lvl.unmarshalText(text) {
		return fmt.Errorf("unrecognized level: %s", text)
	}
	l.SetLevel(_lvl)
	return nil
}

func fromZapLevel(lvl zapcore.Level) Level {
	switch lvl {
	case zapcore.DebugLevel:
		return DebugLevel
	case zapcore.InfoLevel:
		return InfoLevel
	case zapcore.WarnLevel:
		return WarnLevel
	case zapcore.ErrorLevel:
		return ErrorLevel
	case zapcore.DPanicLevel:
		return DPanicLevel
	case zap.PanicLevel:
		return PanicLevel
	case zap.FatalLevel:
		return FatalLevel
	}
	if lvl < zapcore.DebugLevel {
		return TraceLevel
	}
	return FatalLevel
}

func detectLevel(message string) (Level, bool) {
	switch message[1] {
	case 'T':
		if strings.HasPrefix(message, tracePrefix) {
			return TraceLevel, true
		}
	case 'D':
		if strings.HasPrefix(message, debugPrefix) {
			return DebugLevel, true
		}
	case 'I':
		if strings.HasPrefix(message, infoPrefix) {
			return InfoLevel, true
		}
	case 'N':
		if strings.HasPrefix(message, noticePrefix) {
			return NoticeLevel, true
		}
	case 'W':
		if strings.HasPrefix(message, warnPrefix) {
			return WarnLevel, true
		}
	case 'E':
		if strings.HasPrefix(message, errorPrefix) {
			return ErrorLevel, true
		}
	case 'C':
		if strings.HasPrefix(message, criticalPrefix) {
			return CriticalLevel, true
		}
	case 'F':
		if strings.HasPrefix(message, fatalPrefix) {
			return FatalLevel, true
		}
	}
	return 0, false
}
