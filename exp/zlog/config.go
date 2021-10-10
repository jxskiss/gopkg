package zlog

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const defaultLogMaxSize = 300 // MB

const (
	MethodKey = "method"
)

// FileLogConfig serializes file log related config in json/yaml.
type FileLogConfig struct {
	// Filename is the file to write logs to, leave empty to disable file log.
	Filename string `json:"filename" yaml:"filename"`

	// MaxSize is the maximum size in MB of the log file before it gets
	// rotated. It defaults to 300 MB.
	MaxSize int `json:"maxSize" yaml:"maxSize"`

	// MaxDays is the maximum days to retain old log files based on the
	// timestamp encoded in their filenames. The default is not to remove
	// old log files.
	MaxDays int `json:"maxDays" yaml:"maxDays"`

	// MaxBackups is the maximum number of old log files to retain.
	MaxBackups int `json:"maxBackups" yaml:"maxBackups"`
}

// Config serializes log related config in json/yaml.
type Config struct {
	// Level sets the minimum enabled logging level.
	Level string `json:"level" yaml:"level"`

	// Format sets the logger's encoding format.
	// Valid values are "json", "console", and "logfmt".
	Format string `json:"format" yaml:"format"`

	// File specifies file log config.
	File FileLogConfig `json:"file" yaml:"file"`

	// FunctionKey enables logging the function name. By default, function
	// name is not logged.
	FunctionKey string `json:"functionKey" yaml:"functionKey"`

	// Development puts the logger in development mode, which changes the
	// behavior of DPanicLevel and takes stacktraces more liberally.
	Development bool `json:"development" yaml:"development"`

	// DisableTimestamp disables automatic timestamps in output.
	DisableTimestamp bool `json:"disableTimestamp" yaml:"disableTimestamp"`

	// DisableCaller stops annotating logs with the calling function's file
	// name and line number. By default, all logs are annotated.
	DisableCaller bool `json:"disableCaller" yaml:"disableCaller"`

	// DisableStacktrace completely disables automatic stacktrace capturing. By
	// default, stacktraces are captured for WarnLevel and above logs in
	// development and ErrorLevel and above in production.
	DisableStacktrace bool `json:"disableStacktrace" yaml:"disableStacktrace"`

	// Sampling sets a sampling strategy for the logger. Sampling caps the
	// global CPU and I/O load that logging puts on your process while
	// attempting to preserve a representative subset of your logs.
	//
	// Values configured here are per-second. See zapcore.NewSampler for details.
	Sampling *zap.SamplingConfig `json:"sampling" yaml:"sampling"`

	// CtxFunc gets additional logging information from ctx, it's optional.
	// It works only with SetupGlobals, there's no obvious way to support
	// different CtxFunc for non-global individual loggers.
	//
	// See also CtxArgs, CtxResult, WithCtx and Builder.Ctx.
	CtxFunc func(ctx context.Context, args CtxArgs) CtxResult `json:"-" yaml:"-"`

	// Hooks registers functions which will be called each time the Logger
	// writes out an Entry. Repeated use of Hooks is additive.
	//
	// This offers users an easy way to register simple callbacks (e.g.,
	// metrics collection) without implementing the full Core interface.
	//
	// See zap.Hooks and zapcore.RegisterHooks for details.
	Hooks []func(entry zapcore.Entry) error `json:"-" yaml:"-"`
}

// CtxArgs holds arguments passed to Config.CtxFunc.
type CtxArgs struct{}

// CtxResult holds values returned by Config.CtxFunc, which will be used
// to customize a logger's behavior.
type CtxResult struct {
	// Fields will be added to the logger as additional fields.
	Fields []zap.Field

	// An optional Level can be used to dynamically change the logging level.
	Level *Level
}

func (cfg *Config) fillDefaults() {
	if cfg.Level == "" {
		if cfg.Development {
			cfg.Level = "debug"
		} else {
			cfg.Level = "info"
		}
	}
	if cfg.Format == "" {
		if cfg.Development {
			cfg.Format = "console"
		} else {
			cfg.Format = "json"
		}
	}
	if cfg.File.Filename != "" && cfg.File.MaxSize == 0 {
		cfg.File.MaxSize = defaultLogMaxSize
	}
}

func (cfg *Config) buildEncoder() (zapcore.Encoder, error) {
	encConfig := zap.NewProductionEncoderConfig()
	if cfg.Development {
		encConfig = zap.NewDevelopmentEncoderConfig()
	}
	encConfig.FunctionKey = cfg.FunctionKey
	if cfg.DisableTimestamp {
		encConfig.TimeKey = zapcore.OmitKey
	}
	switch cfg.Format {
	case "json":
		return zapcore.NewJSONEncoder(encConfig), nil
	case "console":
		return zapcore.NewConsoleEncoder(encConfig), nil
	case "logfmt":
		return NewLogfmtEncoder(encConfig), nil
	default:
		return nil, fmt.Errorf("unknown format: %s", cfg.Format)
	}
}

func (cfg *Config) buildFileLogger() (*lumberjack.Logger, error) {
	fc := cfg.File
	if st, err := os.Stat(fc.Filename); err == nil {
		if st.IsDir() {
			return nil, errors.New("can't use directory as log filename")
		}
	}
	return &lumberjack.Logger{
		Filename:   fc.Filename,
		MaxSize:    fc.MaxSize,
		MaxAge:     fc.MaxDays,
		MaxBackups: fc.MaxBackups,
		LocalTime:  true,
		Compress:   true,
	}, nil
}

func (cfg *Config) buildOptions(errSink zapcore.WriteSyncer) []zap.Option {
	opts := []zap.Option{zap.ErrorOutput(errSink)}

	if cfg.Development {
		opts = append(opts, zap.Development())
	}

	if !cfg.DisableCaller {
		opts = append(opts, zap.AddCaller())
	}

	if !cfg.DisableStacktrace {
		stackLevel := zap.ErrorLevel
		if cfg.Development {
			stackLevel = zap.WarnLevel
		}
		opts = append(opts, zap.AddStacktrace(stackLevel))
	}

	if cfg.Sampling != nil {
		opts = append(opts, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSampler(core, time.Second, cfg.Sampling.Initial, cfg.Sampling.Thereafter)
		}))
	}
	return opts
}

// NewLogger initializes a zap logger.
func NewLogger(cfg *Config, opts ...zap.Option) (*zap.Logger, *Properties, error) {
	if cfg == nil {
		cfg = &Config{}
	}
	cfg.fillDefaults()
	var output zapcore.WriteSyncer
	if len(cfg.File.Filename) > 0 {
		out, err := cfg.buildFileLogger()
		if err != nil {
			return nil, nil, err
		}
		output = zapcore.AddSync(out)
	} else {
		stderr, _, err := zap.Open("stderr")
		if err != nil {
			return nil, nil, err
		}
		output = stderr
	}
	return NewLoggerWithSyncer(cfg, output, opts...)
}

// NewLoggerWithSyncer initializes a zap logger with given write syncer.
func NewLoggerWithSyncer(cfg *Config, output zapcore.WriteSyncer, opts ...zap.Option) (*zap.Logger, *Properties, error) {
	if cfg == nil {
		cfg = &Config{}
	}
	cfg.fillDefaults()
	encoder, err := cfg.buildEncoder()
	if err != nil {
		return nil, nil, err
	}

	level := newAtomicLevel()
	err = level.UnmarshalText([]byte(cfg.Level))
	if err != nil {
		return nil, nil, err
	}

	// base core at trace level
	core := zapcore.NewCore(encoder, output, TraceLevel)
	if len(cfg.Hooks) > 0 {
		core = zapcore.RegisterHooks(core, cfg.Hooks...)
	}
	opts = append(cfg.buildOptions(output), opts...)
	// wrap the base core with dynamic level
	opts = append(opts, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return &dynamicLevelCore{
			Core:  core,
			level: level.zl,
		}
	}))
	lg := zap.New(core, opts...)
	prop := &Properties{
		cfg:    cfg,
		level:  level,
		core:   core,
		syncer: output,
	}
	return lg, prop, nil
}
