package zlog

import "go.uber.org/zap/zapcore"

type dynamicLevelCore struct {
	zapcore.Core
	baseLevel zapcore.LevelEnabler
	dynLevel  *Level
	levelFunc perLoggerLevelFunc
}

func (c *dynamicLevelCore) Enabled(level zapcore.Level) bool {
	// Dynamic level takes higher priority.
	if c.dynLevel != nil {
		return c.dynLevel.Enabled(level)
	}
	// If per logger level func is configured, leave the filtering work to c.Check.
	if c.levelFunc != nil {
		return true
	}
	return c.baseLevel.Enabled(level)
}

func (c *dynamicLevelCore) With(fields []zapcore.Field) zapcore.Core {
	clone := c.clone()
	clone.Core = clone.Core.With(fields)
	return clone
}

func (c *dynamicLevelCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	var entryLevel = fromZapLevel(entry.Level)
	if len(entry.Message) >= levelPrefixMinLen && entry.Message[0] == '[' {
		if level, detected := detectLevel(entry.Message); detected {
			entryLevel = level
			entry.Level = entryLevel.toZapLevel()
		}
	}

	// Dynamic level takes higher priority.
	if c.dynLevel != nil {
		if *c.dynLevel > entryLevel {
			return ce
		}
		return c.Core.Check(entry, ce)
	}

	// Check per logger levels.
	if c.levelFunc != nil && entry.LoggerName != "" {
		level, found := c.levelFunc(entry.LoggerName)
		if found {
			if level > entryLevel {
				return ce
			}
			return c.Core.Check(entry, ce)
		}
	}

	// Check the configured base level.
	if !c.baseLevel.Enabled(entry.Level) {
		return ce
	}
	return c.Core.Check(entry, ce)
}

func (c *dynamicLevelCore) clone() *dynamicLevelCore {
	return &dynamicLevelCore{
		Core:      c.Core,
		baseLevel: c.baseLevel,
		dynLevel:  c.dynLevel,
		levelFunc: c.levelFunc,
	}
}

func (c *dynamicLevelCore) changeLevel(level Level) *dynamicLevelCore {
	if c.dynLevel != nil && *c.dynLevel == level {
		return c
	}

	clone := c.clone()
	clone.dynLevel = &level
	return clone
}

func tryChangeLevel(level Level) func(zapcore.Core) zapcore.Core {
	return func(core zapcore.Core) zapcore.Core {
		if dyn, ok := core.(*dynamicLevelCore); ok {
			return dyn.changeLevel(level)
		}
		return &dynamicLevelCore{
			Core:      core,
			baseLevel: level,
			dynLevel:  &level,
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