package zlog

import (
	"fmt"
	"strings"

	"go.uber.org/zap/zapcore"
)

type dynamicLevelCore struct {
	zapcore.Core
	baseLevel zapcore.LevelEnabler
	dynLevel  *Level
	levelFunc perLoggerLevelFunc
}

type perLoggerLevelFunc func(name string) (Level, bool)

func buildPerLoggerLevelFunc(levelRules []string) (perLoggerLevelFunc, error) {
	if len(levelRules) == 0 {
		return nil, nil
	}
	tree := &radixTree[Level]{}
	for _, rule := range levelRules {
		tmp := strings.Split(rule, "=")
		if len(tmp) != 2 {
			return nil, fmt.Errorf("invalid per logger level rule: %s", rule)
		}
		loggerName, levelName := tmp[0], tmp[1]
		var level Level
		if !unmarshalLevel(&level, levelName) {
			return nil, fmt.Errorf("unrecognized level: %s", levelName)
		}
		tree.root.insert(loggerName, level)
	}
	return tree.search, nil
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
	var entryLevel = entry.Level

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

func changeLevel(level Level) func(zapcore.Core) zapcore.Core {
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
