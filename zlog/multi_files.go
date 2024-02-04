package zlog

import (
	"go.uber.org/multierr"
	"go.uber.org/zap/zapcore"
)

func newMultiFilesCore(cfg *Config, enc zapcore.Encoder, enab zapcore.LevelEnabler) (
	core *multiFilesCore, closers []func(), err error) {
	core = &multiFilesCore{
		LevelEnabler: enab,
		enc:          enc,
	}
	closers, err = core.initFileWriters(cfg)
	if err != nil {
		return nil, nil, err
	}
	return
}

type multiFilesCore struct {
	zapcore.LevelEnabler
	enc        zapcore.Encoder
	defaultOut zapcore.WriteSyncer
	outFunc    func(string) (zapcore.WriteSyncer, bool)
	outList    []zapcore.WriteSyncer
}

func (c *multiFilesCore) initFileWriters(cfg *Config) (closers []func(), err error) {
	// Close the opened files in case of error occurs.
	defer func() {
		if err != nil {
			runClosers(closers)
			closers = nil
		}
	}()

	var closer func()
	c.defaultOut, closer, err = cfg.FileWriterFactory(&cfg.File)
	if err != nil {
		return closers, err
	}
	closers = append(closers, closer)
	if len(cfg.PerLoggerFiles) == 0 {
		return closers, nil
	}

	tree := &radixTree[zapcore.WriteSyncer]{}
	seenOut := make(map[string]zapcore.WriteSyncer)
	var outList []zapcore.WriteSyncer
	for loggerName, fc := range cfg.PerLoggerFiles {
		if fc.Filename == "" {
			continue
		}
		if fc.Filename == cfg.File.Filename {
			continue
		}
		var out zapcore.WriteSyncer
		if out = seenOut[fc.Filename]; out == nil {
			fc := mergeFileConfig(fc, cfg.File)
			out, closer, err = cfg.FileWriterFactory(&fc)
			if err != nil {
				return closers, err
			}
			seenOut[fc.Filename] = out
			outList = append(outList, out)
			closers = append(closers, closer)
		}
		tree.root.insert(loggerName, out)
	}
	c.outFunc = tree.search
	c.outList = outList
	return closers, nil
}

func (c *multiFilesCore) clone() *multiFilesCore {
	return &multiFilesCore{
		LevelEnabler: c.LevelEnabler,
		enc:          c.enc.Clone(),
		defaultOut:   c.defaultOut,
		outFunc:      c.outFunc,
		outList:      c.outList,
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
	out := c.defaultOut
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
	retErr := c.defaultOut.Sync()
	for _, out := range c.outList {
		err := out.Sync()
		if err != nil {
			retErr = multierr.Append(retErr, err)
		}
	}
	return retErr
}
