package zlog

import "go.uber.org/zap/zapcore"

type dynamicLevelCore struct {
	zapcore.Core
	level zapcore.LevelEnabler
}

func (c *dynamicLevelCore) Enabled(level zapcore.Level) bool {
	return c.level.Enabled(level)
}

func (c *dynamicLevelCore) With(fields []zapcore.Field) zapcore.Core {
	return &dynamicLevelCore{c.Core.With(fields), c.level}
}

func (c *dynamicLevelCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if !c.Enabled(entry.Level) {
		return ce
	}
	return c.Core.Check(entry, ce)
}

func (c *dynamicLevelCore) changeLevel(level Level) *dynamicLevelCore {
	if c.level == level {
		return c
	}
	return &dynamicLevelCore{
		Core:  c.Core,
		level: level,
	}
}

func tryChangeLevel(level Level) func(zapcore.Core) zapcore.Core {
	return func(core zapcore.Core) zapcore.Core {
		dyn, ok := core.(*dynamicLevelCore)
		if ok {
			if dyn.level == level {
				return core
			}
			return dyn.changeLevel(level)
		}
		return &dynamicLevelCore{
			Core:  core,
			level: level,
		}
	}
}

// for testing
func unwrapDynamicLevelCore(core zapcore.Core) zapcore.Core {
	for {
		wrapped, ok := core.(*dynamicLevelCore)
		if !ok {
			return core
		}
		core = wrapped.Core
	}
}
