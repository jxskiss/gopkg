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

const defaultMethodNameKey = "methodName"

// FileLogConfig serializes file log related config in json/yaml.
type FileLogConfig struct {
	// Filename is the file to write logs to, leave empty to disable file log.
	Filename string `json:"filename" yaml:"filename"`

	// MaxSize is the maximum size in MB of the log file before it gets
	// rotated. It defaults to 100 MB.
	MaxSize int `json:"maxSize" yaml:"maxSize"`

	// MaxDays is the maximum days to retain old log files based on the
	// timestamp encoded in their filenames.
	// The default is not to remove old log files.
	MaxDays int `json:"maxDays" yaml:"maxDays"`

	// MaxBackups is the maximum number of old log files to retain.
	MaxBackups int `json:"maxBackups" yaml:"maxBackups"`

	// LocalTime determines if the time used for formatting the timestamps in
	// backup files is the computer's local time.
	// The default is to use UTC time.
	LocalTime bool `json:"localTime" yaml:"localTime"`

	// Compress determines if the rotated log files should be compressed
	// using gzip. The default is not to perform compression.
	Compress bool `json:"compress" yaml:"compress"`
}

// GlobalConfig configures some global behavior of this package.
type GlobalConfig struct {
	// RedirectStdLog redirects output from the standard log library's
	// package-global logger to the global logger in this package at
	// InfoLevel.
	RedirectStdLog bool `json:"redirectStdLog" yaml:"redirectStdLog"`

	// DisableTrace disables trace level messages.
	//
	// Disabling trace level messages makes the trace logging functions no-op,
	// it gives better performance when you definitely don't need TraceLevel
	// messages (e.g. in production deployment).
	DisableTrace bool `json:"disableTrace" yaml:"disableTrace"`

	// MethodNameKey specifies the key to use when adding caller's method
	// name to logging messages. It defaults to "methodName".
	MethodNameKey string `json:"methodNameKey" yaml:"methodNameKey"`

	// CtxFunc gets additional logging information from ctx, it's optional.
	//
	// See also CtxArgs, CtxResult, WithCtx and Builder.Ctx.
	CtxFunc CtxFunc `json:"-" yaml:"-"`
}

// Config serializes log related config in json/yaml.
type Config struct {
	// Level sets the default logging level for the logger.
	Level string `json:"level" yaml:"level"`

	// PerLoggerLevels optionally configures logging level by logger names.
	// The format is "loggerName.subLogger=level".
	// If a level is configured for a parent logger, but not configured for
	// a child logger, the child logger derives from its parent.
	PerLoggerLevels []string `json:"perLoggerLevels" yaml:"perLoggerLevels"`

	// Format sets the logger's encoding format.
	// Valid values are "json", "console", and "logfmt".
	Format string `json:"format" yaml:"format"`

	// File specifies file log config.
	File FileLogConfig `json:"file" yaml:"file"`

	// PerLoggerFiles optionally set different file destination for different
	// loggers specified by logger name.
	// If a destination is configured for a parent logger, but not configured
	// for a child logger, the child logger derives from its parent.
	PerLoggerFiles map[string]FileLogConfig `json:"perLoggerFiles" yaml:"perLoggerFiles"`

	// FunctionKey enables logging the function name. By default, function
	// name is not logged.
	FunctionKey string `json:"functionKey" yaml:"functionKey"`

	// Development puts the logger in development mode, which changes the
	// behavior of DPanicLevel and takes stacktrace more liberally.
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

	// Sampling sets a sampling strategy for the logger. Sampling caps the
	// global CPU and I/O load that logging puts on your process while
	// attempting to preserve a representative subset of your logs.
	//
	// Values configured here are per-second. See zapcore.NewSampler for details.
	Sampling *zap.SamplingConfig `json:"sampling" yaml:"sampling"`

	// Hooks registers functions which will be called each time the Logger
	// writes out an Entry. Repeated use of Hooks is additive.
	//
	// This offers users an easy way to register simple callbacks (e.g.,
	// metrics collection) without implementing the full Core interface.
	//
	// See zap.Hooks and zapcore.RegisterHooks for details.
	Hooks []func(zapcore.Entry) error `json:"-" yaml:"-"`

	// GlobalConfig configures some global behavior of this package.
	// It works with SetupGlobals and ReplaceGlobals, it has no effect for
	// non-global individual loggers.
	GlobalConfig `yaml:",inline"`
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

func (cfg *Config) fillDefaults() *Config {
	if cfg == nil {
		cfg = &Config{}
	}
	if cfg.Level == "" {
		if cfg.Development {
			cfg.Level = "trace"
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
	return cfg
}

func (cfg *Config) buildEncoder() (zapcore.Encoder, error) {
	encConfig := zap.NewProductionEncoderConfig()
	encConfig.EncodeLevel = func(lv zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(fromZapLevel(lv).String())
	}
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
		encConfig.EncodeLevel = func(lv zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(fromZapLevel(lv).CapitalString())
		}
		return zapcore.NewConsoleEncoder(encConfig), nil
	case "logfmt":
		return NewLogfmtEncoder(encConfig), nil
	default:
		return nil, fmt.Errorf("unknown format: %s", cfg.Format)
	}
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
//
// If Config.File is configured, the log messages will be written to the
// specified file with rotation, else they will be written to stderr.
//
// The returned zap.Logger supports dynamic level, see Config.PerLoggerLevels
// and GlobalConfig.CtxFunc for details about dynamic level.
// The returned zap.Logger and Properties may be passed to ReplaceGlobals
// to change the global logger and customize some global behavior of this
// package.
func New(cfg *Config, opts ...zap.Option) (*zap.Logger, *Properties, error) {
	cfg = cfg.fillDefaults()
	var err error
	var output zapcore.WriteSyncer
	if len(cfg.File.Filename) > 0 {
		if len(cfg.PerLoggerFiles) > 0 {
			return newWithMultiFilesOutput(cfg, opts...)
		}
		output, err = buildFileLogger(cfg.File)
		if err != nil {
			return nil, nil, err
		}
	} else {
		output, _, err = zap.Open("stderr")
		if err != nil {
			return nil, nil, err
		}
	}
	return NewWithOutput(cfg, output, opts...)
}

func newWithMultiFilesOutput(cfg *Config, opts ...zap.Option) (*zap.Logger, *Properties, error) {
	enc, err := cfg.buildEncoder()
	if err != nil {
		return nil, nil, err
	}

	// base multi-files core at trace level
	core, err := newMultiFilesCore(cfg, enc, TraceLevel)
	if err != nil {
		return nil, nil, err
	}

	aLevel := newAtomicLevel()
	err = aLevel.UnmarshalText([]byte(cfg.Level))
	if err != nil {
		return nil, nil, err
	}

	cfgOpts, err := cfg.buildOptions()
	if err != nil {
		return nil, nil, err
	}
	opts = append(cfgOpts, opts...)

	wcc := &WrapCoreConfig{
		Level:           0, // not used here
		PerLoggerLevels: cfg.PerLoggerLevels,
		Hooks:           cfg.Hooks,
		GlobalConfig:    cfg.GlobalConfig,
	}
	return newWithWrapCoreConfig(wcc, core, aLevel, opts...)
}

// NewWithOutput initializes a zap logger with given write syncer as output
// destination.
//
// The returned zap.Logger supports dynamic level, see Config.PerLoggerLevels
// and GlobalConfig.CtxFunc for details about dynamic level.
// The returned zap.Logger and Properties may be passed to ReplaceGlobals
// to change the global logger and customize some global behavior of this
// package.
func NewWithOutput(cfg *Config, output zapcore.WriteSyncer, opts ...zap.Option) (*zap.Logger, *Properties, error) {
	cfg = cfg.fillDefaults()
	encoder, err := cfg.buildEncoder()
	if err != nil {
		return nil, nil, err
	}

	// base core at trace level
	core := zapcore.NewCore(encoder, output, TraceLevel)

	aLevel := newAtomicLevel()
	err = aLevel.UnmarshalText([]byte(cfg.Level))
	if err != nil {
		return nil, nil, err
	}

	cfgOpts, err := cfg.buildOptions()
	if err != nil {
		return nil, nil, err
	}
	opts = append(cfgOpts, opts...)

	wcc := &WrapCoreConfig{
		Level:           0, // not used here
		PerLoggerLevels: cfg.PerLoggerLevels,
		Hooks:           cfg.Hooks,
		GlobalConfig:    cfg.GlobalConfig,
	}
	return newWithWrapCoreConfig(wcc, core, aLevel, opts...)
}

type WrapCoreConfig struct {
	// Level sets the default logging level for the logger.
	Level Level

	// PerLoggerLevels optionally configures logging level by logger names.
	// The format is "loggerName.subLogger=level".
	// If a level is configured for a parent logger, but not configured for
	// a child logger, the child logger will derive the level from its parent.
	PerLoggerLevels []string

	// Hooks registers functions which will be called each time the Logger
	// writes out an Entry. Repeated use of Hooks is additive.
	//
	// This offers users an easy way to register simple callbacks (e.g.,
	// metrics collection) without implementing the full Core interface.
	//
	// See zap.Hooks and zapcore.RegisterHooks for details.
	Hooks []func(zapcore.Entry) error

	// GlobalConfig configures some global behavior of this package.
	// It works with SetupGlobals and ReplaceGlobals, it has no effect for
	// non-global individual loggers.
	GlobalConfig `yaml:",inline"`
}

// NewWithCore initializes a zap logger with given core.
//
// You may use this function to integrate with custom cores (e.g. to
// integrate with Sentry or Graylog, or output to multiple sinks).
//
// The returned zap.Logger supports dynamic level, see
// WrapCoreConfig.PerLoggerLevels and GlobalConfig.CtxFunc for details
// about dynamic level. Note that if you want to use the dynamic level
// feature, the provided core must be configured to log low level messages
// (e.g. debug).
//
// The returned zap.Logger and Properties may be passed to ReplaceGlobals
// to change the global logger and customize some global behavior of this
// package.
func NewWithCore(cfg *WrapCoreConfig, core zapcore.Core, opts ...zap.Option) (*zap.Logger, *Properties, error) {
	if cfg == nil {
		cfg = &WrapCoreConfig{Level: InfoLevel}
	}

	aLevel := newAtomicLevel()
	aLevel.SetLevel(cfg.Level)

	return newWithWrapCoreConfig(cfg, core, aLevel, opts...)
}

func newWithWrapCoreConfig(
	cfg *WrapCoreConfig,
	core zapcore.Core,
	aLevel atomicLevel,
	opts ...zap.Option,
) (*zap.Logger, *Properties, error) {
	if len(cfg.Hooks) > 0 {
		core = zapcore.RegisterHooks(core, cfg.Hooks...)
	}

	// build per logger level rules
	perLoggerLevelFn, err := buildPerLoggerLevelFunc(cfg.PerLoggerLevels)
	if err != nil {
		return nil, nil, err
	}

	// wrap the base core with dynamic level
	opts = append(opts, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return &dynamicLevelCore{
			Core:      core,
			baseLevel: aLevel.zl,
			levelFunc: perLoggerLevelFn,
		}
	}))

	lg := zap.New(core, opts...)
	prop := &Properties{
		cfg:   cfg.GlobalConfig,
		level: aLevel,
	}
	return lg, prop, nil
}

func buildFileLogger(fc FileLogConfig) (zapcore.WriteSyncer, error) {
	if st, err := os.Stat(fc.Filename); err == nil {
		if st.IsDir() {
			return nil, errors.New("cannot use directory as log filename")
		}
	}
	out := &lumberjack.Logger{
		Filename:   fc.Filename,
		MaxSize:    fc.MaxSize,
		MaxAge:     fc.MaxDays,
		MaxBackups: fc.MaxBackups,
		LocalTime:  fc.LocalTime,
		Compress:   fc.Compress,
	}
	ws := zapcore.AddSync(out)
	return ws, nil
}

func mergeFileLogConfig(fc FileLogConfig, deft FileLogConfig) FileLogConfig {
	setIfZero(&fc.MaxSize, deft.MaxSize)
	setIfZero(&fc.MaxDays, deft.MaxDays)
	setIfZero(&fc.MaxBackups, deft.MaxBackups)
	setIfZero(&fc.LocalTime, deft.LocalTime)
	setIfZero(&fc.Compress, deft.Compress)
	return fc
}

func setIfZero[T comparable](dst *T, value T) {
	var zero T
	if *dst == zero {
		*dst = value
	}
}
