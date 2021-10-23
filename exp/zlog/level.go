package zlog

import (
	"fmt"

	"go.uber.org/atomic"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Level int8

const (
	TraceLevel Level = iota
	DebugLevel
	InfoLevel

	// NoticeLevel is not implemented.
	NoticeLevel

	WarnLevel
	ErrorLevel

	// CriticalLevel is not implemented.
	CriticalLevel

	DPanicLevel
	PanicLevel
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

func (l Level) toZapLevel() zapcore.Level { return mapZapLevels[l] }

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

func getLevelFromEnabler(enab zapcore.LevelEnabler) Level {
	for i, lvl := range mapZapLevels {
		if enab.Enabled(lvl) {
			return Level(i)
		}
	}
	return InfoLevel
}
