package zlog

import (
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/jxskiss/gopkg/v2/zlog/internal/terminal"
)

const (
	defaultMethodNameKey = "methodName"
	consoleTimeLayout    = "2006/01/02 15:04:05.000000"
)

// FileConfig serializes file log related config in json/yaml.
type FileConfig struct {
	// Filename is the file to write logs to, leave empty to disable file log.
	Filename string `json:"filename" yaml:"filename"`

	// MaxSize is the maximum size in MB of the log file before it gets
	// rotated. It defaults to 100 MB.
	MaxSize int `json:"maxSize" yaml:"maxSize"`

	// MaxDays is the maximum days to retain old log files based on the
	// timestamp encoded in their filenames.
	// Note that a day is defined as 24 hours and may not exactly correspond
	// to calendar days due to daylight savings, leap seconds, etc.
	// The default is not to remove old log files.
	MaxDays int `json:"maxDays" yaml:"maxDays"`

	// MaxBackups is the maximum number of old log files to retain.
	// The default is to retain all old log files (though MaxAge may still
	// cause them to get deleted.)
	MaxBackups int `json:"maxBackups" yaml:"maxBackups"`

	// Compress determines if the rotated log files should be compressed.
	// The default is not to perform compression.
	Compress bool `json:"compress" yaml:"compress"`
}

// FileWriterFactory opens a file to write log, FileConfig specifies
// filename and optional settings to rotate the log files.
// The returned WriteSyncer should be safe for concurrent use,
// you may use zap.Lock to wrap a WriteSyncer to be concurrent safe.
// It also returns any error encountered and a function to close
// the opened file.
//
// User may check github.com/jxskiss/gopkg/_examples/zlog/lumberjack_writer
// for an example to use "lumberjack.v2" as a rolling logger.
type FileWriterFactory func(fc *FileConfig) (zapcore.WriteSyncer, func(), error)

// GlobalConfig configures some global behavior of this package.
type GlobalConfig struct {
	// RedirectStdLog redirects output from the standard log library's
	// package-global logger to the global logger in this package at
	// InfoLevel.
	RedirectStdLog bool `json:"redirectStdLog" yaml:"redirectStdLog"`

	// MethodNameKey specifies the key to use when adding caller's method
	// name to logging messages. It defaults to "methodName".
	MethodNameKey string `json:"methodNameKey" yaml:"methodNameKey"`

	// TraceFilterRule optionally configures filter rule to allow or deny
	// trace logging in some packages or files.
	//
	// It uses glob to match filename, the syntax is "allow=glob1,glob2;deny=glob3,glob4".
	// For example:
	//
	// - "", empty rule means allow all messages
	// - "allow=all", allow all messages
	// - "deny=all", deny all messages
	// - "allow=pkg1/*,pkg2/*.go",
	//   allow messages from files in `pkg1` and `pkg2`,
	//   deny messages from all other packages
	// - "allow=pkg1/sub1/abc.go,pkg1/sub2/def.go",
	//   allow messages from file `pkg1/sub1/abc.go` and `pkg1/sub2/def.go`,
	//   deny messages from all other files
	// - "allow=pkg1/**",
	//   allow messages from files and sub-packages in `pkg1`,
	//   deny messages from all other packages
	// - "deny=pkg1/**.go,pkg2/**.go",
	//   deny messages from files and sub-packages in `pkg1` and `pkg2`,
	//   allow messages from all other packages
	// - "allow=all;deny=pkg/**", same as "deny=pkg/**"
	//
	// If both "allow" and "deny" directives are configured, the "allow" directive
	// takes effect, the "deny" directive is ignored.
	//
	// The default value is empty, which means all messages are allowed.
	//
	// User can also set the environment variable "ZLOG_TRACE_FILTER_RULE"
	// to configure it at runtime, if available, the environment variable
	// is used when this value is empty.
	TraceFilterRule string `json:"traceFilterRule" yaml:"traceFilterRule"`

	// CtxHandler customizes a logger's behavior at runtime dynamically.
	CtxHandler CtxHandler `json:"-" yaml:"-"`
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
	File FileConfig `json:"file" yaml:"file"`

	// PerLoggerFiles optionally set different file destination for different
	// loggers specified by logger name.
	// If a destination is configured for a parent logger, but not configured
	// for a child logger, the child logger derives from its parent.
	PerLoggerFiles map[string]FileConfig `json:"perLoggerFiles" yaml:"perLoggerFiles"`

	// FileWriterFactory optionally specifies a custom factory function,
	// when File is configured, to open a file to write log.
	// By default, [zap.Open] is used, which does not support file rotation.
	FileWriterFactory FileWriterFactory `json:"-" yaml:"-"`

	// FunctionKey enables logging the function name.
	// By default, function name is not logged.
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
	// By default, stacktraces are captured for ErrorLevel and above.
	StacktraceLevel string `json:"stacktraceLeve" yaml:"stacktraceLevel"`

	// Sampling sets a sampling strategy for the logger. Sampling caps the
	// global CPU and I/O load that logging puts on your process while
	// attempting to preserve a representative subset of your logs.
	//
	// Values configured here are per-second. See zapcore.NewSampler for details.
	Sampling *zap.SamplingConfig `json:"sampling" yaml:"sampling"`

	// Hooks registers functions which will be called each time the logger
	// writes out an Entry. Repeated use of Hooks is additive.
	//
	// This offers users an easy way to register simple callbacks (e.g.,
	// metrics collection) without implementing the full Core interface.
	//
	// See zap.Hooks and zapcore.RegisterHooks for details.
	Hooks []func(zapcore.Entry) error `json:"-" yaml:"-"`

	// GlobalConfig configures some global behavior of this package.
	// It works with SetupGlobals and ReplaceGlobals, it has no effect for
	// individual non-global loggers.
	GlobalConfig `yaml:",inline"`
}

func (cfg *Config) checkAndFillDefaults() *Config {
	if cfg == nil {
		cfg = &Config{}
	}
	if cfg.FileWriterFactory == nil {
		cfg.FileWriterFactory = func(fc *FileConfig) (zapcore.WriteSyncer, func(), error) {
			return zap.Open(fc.Filename)
		}
	}
	if cfg.Development {
		setIfZero(&cfg.Level, "trace")
		setIfZero(&cfg.Format, "console")
	} else {
		setIfZero(&cfg.Level, "info")
		setIfZero(&cfg.Format, "json")
	}
	setIfZero(&cfg.StacktraceLevel, "error")
	return cfg
}

func (cfg *Config) buildEncoder(isStderr bool) (zapcore.Encoder, error) {
	encConfig := zap.NewProductionEncoderConfig()
	encConfig.EncodeLevel = encodeLevelLowercase
	if cfg.Development {
		encConfig = zap.NewDevelopmentEncoderConfig()
		encConfig.EncodeLevel = encodeLevelCapital
	}
	encConfig.FunctionKey = cfg.FunctionKey
	if cfg.DisableTimestamp {
		encConfig.TimeKey = zapcore.OmitKey
	}
	switch cfg.Format {
	case "json":
		return zapcore.NewJSONEncoder(encConfig), nil
	case "console":
		encConfig.EncodeLevel = encodeLevelCapital
		encConfig.EncodeTime = zapcore.TimeEncoderOfLayout(consoleTimeLayout)
		encConfig.ConsoleSeparator = " "
		if isStderr && terminal.CheckIsTerminal(os.Stderr) {
			encConfig.EncodeLevel = encodeLevelColorCapital
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
		if !unmarshalLevel(&stackLevel, cfg.StacktraceLevel) {
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
// If Config.File is configured, logs will be written to the specified file,
// and Config.PerLoggerFiles can be used to write logs to different files
// specified by logger name.
// By default, logs are written to stderr.
//
// The returned zap.Logger supports dynamic level, see Config.PerLoggerLevels
// and GlobalConfig.CtxHandler for details about dynamic level.
// The returned zap.Logger and Properties may be passed to ReplaceGlobals
// to change the global logger and customize some global behavior of this
// package.
func New(cfg *Config, opts ...zap.Option) (*zap.Logger, *Properties, error) {
	cfg = cfg.checkAndFillDefaults()
	var err error
	var output zapcore.WriteSyncer
	var closer func()
	if len(cfg.File.Filename) > 0 {
		if len(cfg.PerLoggerFiles) > 0 {
			return newWithMultiFilesOutput(cfg, opts...)
		}
		output, closer, err = cfg.FileWriterFactory(&cfg.File)
		if err != nil {
			return nil, nil, err
		}
	} else {
		output, closer, err = zap.Open("stderr")
		if err != nil {
			return nil, nil, err
		}
		output = &wrapStderr{output}
	}
	l, p, err := NewWithOutput(cfg, output, opts...)
	if err != nil {
		closer()
		return nil, nil, err
	}
	p.closers = append(p.closers, closer)
	return l, p, nil
}

func newWithMultiFilesOutput(cfg *Config, opts ...zap.Option) (*zap.Logger, *Properties, error) {
	enc, err := cfg.buildEncoder(false)
	if err != nil {
		return nil, nil, err
	}

	var level Level
	if !unmarshalLevel(&level, cfg.Level) {
		return nil, nil, fmt.Errorf("unrecognized level: %s", cfg.Level)
	}

	cfgOpts, err := cfg.buildOptions()
	if err != nil {
		return nil, nil, err
	}
	opts = append(cfgOpts, opts...)

	// base multi-files core at trace level
	core, closers, err := newMultiFilesCore(cfg, enc, TraceLevel)
	if err != nil {
		return nil, nil, err
	}

	wcc := &WrapCoreConfig{
		Level:           level,
		PerLoggerLevels: cfg.PerLoggerLevels,
		Hooks:           cfg.Hooks,
		GlobalConfig:    cfg.GlobalConfig,
	}
	l, p, err := newWithWrapCoreConfig(wcc, core, opts...)
	if err != nil {
		runClosers(closers)
		return nil, nil, err
	}
	p.closers = closers
	return l, p, nil
}

// NewWithOutput initializes a zap logger with given write syncer as output.
//
// The returned zap.Logger supports dynamic level, see Config.PerLoggerLevels
// and GlobalConfig.CtxHandler for details about dynamic level.
// The returned zap.Logger and Properties may be passed to ReplaceGlobals
// to change the global logger and customize some global behavior of this
// package.
func NewWithOutput(cfg *Config, output zapcore.WriteSyncer, opts ...zap.Option) (*zap.Logger, *Properties, error) {
	cfg = cfg.checkAndFillDefaults()

	isStderr := false
	if wrapper, ok := output.(*wrapStderr); ok {
		isStderr = true
		output = wrapper.WriteSyncer
	}
	encoder, err := cfg.buildEncoder(isStderr)
	if err != nil {
		return nil, nil, err
	}

	// base core logging any level messages
	core := zapcore.NewCore(encoder, output, Level(-127))

	var level Level
	if !unmarshalLevel(&level, cfg.Level) {
		return nil, nil, fmt.Errorf("unrecognized level: %s", cfg.Level)
	}

	cfgOpts, err := cfg.buildOptions()
	if err != nil {
		return nil, nil, err
	}
	opts = append(cfgOpts, opts...)

	wcc := &WrapCoreConfig{
		Level:           level,
		PerLoggerLevels: cfg.PerLoggerLevels,
		Hooks:           cfg.Hooks,
		GlobalConfig:    cfg.GlobalConfig,
	}
	return newWithWrapCoreConfig(wcc, core, opts...)
}

type WrapCoreConfig struct {
	// Level sets the default logging level for the logger.
	Level Level `json:"level" yaml:"level"`

	// PerLoggerLevels optionally configures logging level by logger names.
	// The format is "loggerName.subLogger=level".
	// If a level is configured for a parent logger, but not configured for
	// a child logger, the child logger will derive the level from its parent.
	PerLoggerLevels []string `json:"perLoggerLevels" yaml:"perLoggerLevels"`

	// Hooks registers functions which will be called each time the logger
	// writes out an Entry. Repeated use of Hooks is additive.
	//
	// This offers users an easy way to register simple callbacks (e.g.,
	// metrics collection) without implementing the full Core interface.
	//
	// See zap.Hooks and zapcore.RegisterHooks for details.
	Hooks []func(zapcore.Entry) error `json:"-" yaml:"-"`

	// GlobalConfig configures some global behavior of this package.
	// It works with SetupGlobals and ReplaceGlobals, it has no effect for
	// individual non-global loggers.
	GlobalConfig `yaml:",inline"`
}

// NewWithCore initializes a zap logger with given core.
//
// You may use this function to integrate with custom cores (e.g. to
// integrate with Sentry or Graylog, or output to multiple sinks).
//
// The returned zap.Logger supports dynamic level, see
// WrapCoreConfig.PerLoggerLevels and GlobalConfig.CtxHandler for details
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
	return newWithWrapCoreConfig(cfg, core, opts...)
}

func newWithWrapCoreConfig(
	cfg *WrapCoreConfig,
	core zapcore.Core,
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
	aLevel := zap.NewAtomicLevelAt(cfg.Level)
	opts = append(opts, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return &dynamicLevelCore{
			Core:      core,
			baseLevel: aLevel,
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

func mergeFileConfig(fc, defaultConfig FileConfig) FileConfig {
	setIfZero(&fc.MaxSize, defaultConfig.MaxSize)
	setIfZero(&fc.MaxDays, defaultConfig.MaxDays)
	setIfZero(&fc.MaxBackups, defaultConfig.MaxBackups)
	setIfZero(&fc.Compress, defaultConfig.Compress)
	return fc
}

func setIfZero[T comparable](dst *T, value T) {
	var zero T
	if *dst == zero {
		*dst = value
	}
}

func runClosers(closers []func()) {
	for _, closeFunc := range closers {
		closeFunc()
	}
}

type wrapStderr struct {
	zapcore.WriteSyncer
}
