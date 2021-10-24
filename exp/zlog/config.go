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

const MethodKey = "method"

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

	// DisableStacktrace disables automatic stacktrace capturing.
	DisableStacktrace bool `json:"disableStacktrace" yaml:"disableStacktrace"`

	// StacktraceLevel sets the level that stacktrace will be captured.
	// By default, stacktraces are captured for WarnLevel and above logs in
	// development and ErrorLevel and above in production.
	StacktraceLevel string `json:"stacktraceLeve" yaml:"stacktraceLevel"`

	// UseMilliClock optionally configures the logger to use a low precision
	// clock at milliseconds to optimize heavy logging use case to get best
	// performance. By default, the system clock is used, and in most cases
	// the system clock is good enough.
	UseMilliClock int `json:"useMilliClock" yaml:"useMilliClock"`

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
	CtxFunc CtxFunc `json:"-" yaml:"-"`

	// Hooks registers functions which will be called each time the Logger
	// writes out an Entry. Repeated use of Hooks is additive.
	//
	// This offers users an easy way to register simple callbacks (e.g.,
	// metrics collection) without implementing the full Core interface.
	//
	// See zap.Hooks and zapcore.RegisterHooks for details.
	Hooks []func(zapcore.Entry) error `json:"-" yaml:"-"`
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

// CtxFunc gets additional logging data from ctx, it may return extra fields
// to attach to the logging entry, or change the logging level dynamically.
type CtxFunc func(ctx context.Context, args CtxArgs) CtxResult

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
	if cfg.StacktraceLevel == "" {
		if cfg.Development {
			cfg.StacktraceLevel = "warn"
		} else {
			cfg.StacktraceLevel = "error"
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

func (cfg *Config) buildOptions() ([]zap.Option, error) {
	var opts []zap.Option
	if cfg.Development {
		opts = append(opts, zap.Development())
	}
	if !cfg.DisableCaller {
		opts = append(opts, zap.AddCaller())
	}
	if !cfg.DisableStacktrace {
		var stackLevel Level
		if !stackLevel.unmarshalText([]byte(cfg.StacktraceLevel)) {
			return nil, fmt.Errorf("unrecognized stacktrace level: %s", cfg.StacktraceLevel)
		}
		opts = append(opts, zap.AddStacktrace(stackLevel))
	}
	if cfg.UseMilliClock > 0 {
		opts = append(opts, zap.WithClock(newMilliClock(cfg.UseMilliClock)))
	}
	if cfg.Sampling != nil {
		opts = append(opts, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			tick := time.Second
			first, thereafter := cfg.Sampling.Initial, cfg.Sampling.Thereafter
			return zapcore.NewSamplerWithOptions(core, tick, first, thereafter)
		}))
	}
	return opts, nil
}

// New initializes a zap logger.
func New(cfg *Config, opts ...zap.Option) (*zap.Logger, *Properties, error) {
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
	return NewWithOutput(cfg, output, opts...)
}

// NewWithOutput initializes a zap logger with given write syncer.
func NewWithOutput(cfg *Config, output zapcore.WriteSyncer, opts ...zap.Option) (*zap.Logger, *Properties, error) {
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

	cfgOpts, err := cfg.buildOptions()
	if err != nil {
		return nil, nil, err
	}
	opts = append(cfgOpts, opts...)

	// base core at trace level
	core := zapcore.NewCore(encoder, output, TraceLevel)
	if len(cfg.Hooks) > 0 {
		core = zapcore.RegisterHooks(core, cfg.Hooks...)
	}
	// wrap the base core with dynamic level
	opts = append(opts, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return &dynamicLevelCore{
			Core:  core,
			level: level.zl,
		}
	}))
	lg := zap.New(core, opts...)
	prop := &Properties{
		functionKey: cfg.FunctionKey,
		ctxFunc:     cfg.CtxFunc,
		level:       level,
	}
	return lg, prop, nil
}

// NewWithCore initializes a zap logger with given core and level.
// If you want to use the dynamic level feature, the provided core must be
// configured to logging low level messages.
//
// You may use this function to integrate with custom cores (e.g. to
// integrate with Sentry or Graylog, or output to multiple sinks).
func NewWithCore(
	core zapcore.Core,
	level Level,
	ctxFunc CtxFunc,
	hooks []func(zapcore.Entry) error,
	opts ...zap.Option,
) (*zap.Logger, *Properties, error) {
	atomLevel := newAtomicLevel()
	atomLevel.SetLevel(level)
	opts = append(opts, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return &dynamicLevelCore{
			Core:  core,
			level: atomLevel.zl,
		}
	}))
	if len(hooks) > 0 {
		core = zapcore.RegisterHooks(core, hooks...)
	}
	lg := zap.New(core, opts...)
	prop := &Properties{
		functionKey: "",
		ctxFunc:     ctxFunc,
		level:       atomLevel,
	}
	return lg, prop, nil
}
