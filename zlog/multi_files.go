package zlog

import "go.uber.org/zap/zapcore"

func newMultiFilesCore(cfg *Config, enc zapcore.Encoder, enab zapcore.LevelEnabler) (*multiFilesCore, error) {
	defaultOut, err := buildFileLogger(cfg.File)
	if err != nil {
		return nil, err
	}
	core := &multiFilesCore{
		LevelEnabler: enab,
		enc:          enc,
		deftOut:      defaultOut,
	}
	err = core.buildPerLoggerOutFunc(cfg)
	if err != nil {
		return nil, err
	}
	return core, nil
}

type multiFilesCore struct {
	zapcore.LevelEnabler
	enc     zapcore.Encoder
	deftOut zapcore.WriteSyncer
	outFunc func(string) (zapcore.WriteSyncer, bool)
	outList []zapcore.WriteSyncer
}

func (c *multiFilesCore) buildPerLoggerOutFunc(cfg *Config) error {
	if len(cfg.PerLoggerFiles) == 0 {
		return nil
	}

	tree := &radixTree[zapcore.WriteSyncer]{}
	seenOut := make(map[string]zapcore.WriteSyncer)
	for loggerName, fc := range cfg.PerLoggerFiles {
		if fc.Filename == cfg.File.Filename {
			continue
		}
		var err error
		var out zapcore.WriteSyncer
		if out = seenOut[fc.Filename]; out == nil {
			fc := mergeFileLogConfig(fc, cfg.File)
			out, err = buildFileLogger(fc)
			if err != nil {
				return err
			}
			seenOut[fc.Filename] = out
			c.outList = append(c.outList, out)
		}
		tree.root.insert(loggerName, out)
	}
	c.outFunc = tree.search
	return nil
}

func (c *multiFilesCore) clone() *multiFilesCore {
	return &multiFilesCore{
		LevelEnabler: c.LevelEnabler,
		enc:          c.enc.Clone(),
		deftOut:      c.deftOut,
		outFunc:      c.outFunc,
	}
}

func (c *multiFilesCore) With(fields []zapcore.Field) zapcore.Core {
	clone := c.clone()
	addFields(clone.enc, fields)
	return clone
}

func (c *multiFilesCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

func (c *multiFilesCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	buf, err := c.enc.EncodeEntry(ent, fields)
	if err != nil {
		return err
	}
	out := c.deftOut
	if c.outFunc != nil {
		tmp, found := c.outFunc(ent.LoggerName)
		if found {
			out = tmp
		}
	}
	_, err = out.Write(buf.Bytes())
	buf.Free()
	if err != nil {
		return err
	}
	if ent.Level > zapcore.ErrorLevel {
		// Since we may be crashing the program, sync the output.
		// Ignore Sync errors, pending a clean solution to issue
		// https://github.com/uber-go/zap/issues/370.
		out.Sync()
	}
	return nil
}

func (c *multiFilesCore) Sync() error {
	err := c.deftOut.Sync()
	if err != nil {
		return err
	}
	for _, out := range c.outList {
		err = out.Sync()
		if err != nil {
			return err
		}
	}
	return nil
}
