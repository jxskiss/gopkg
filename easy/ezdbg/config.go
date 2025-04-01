package ezdbg

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jxskiss/gopkg/v2/internal/logfilter"
)

const FilterRuleEnvName = "EZDBG_FILTER_RULE"

// Config configures the behavior of functions in this package.
func Config(cfg Cfg) {
	if cfg.FilterRule == "" {
		envRule := os.Getenv(FilterRuleEnvName)
		if envRule != "" {
			stdLogger{}.Infof("ezdbg: using filter rule from env: %q", envRule)
			cfg.FilterRule = envRule
		}
	}
	if cfg.FilterRule != "" {
		var errs []error
		cfg.filter, errs = logfilter.NewFileNameFilter(cfg.FilterRule)
		for _, err := range errs {
			stdLogger{}.Warnf("ezdbg: %v", err)
		}
	}
	_logcfg = cfg
}

var _logcfg Cfg

// Cfg provides optional config to configure this package.
type Cfg struct {

	// EnableDebug determines whether debug log is enabled, it may use
	// the given context.Context to enable or disable request-level debug log.
	// If EnableDebug returns false, the log message is discarded.
	//
	// User must configure this to enable debug log from this package.
	// By default, functions in this package discard all messages.
	EnableDebug func(context.Context) bool

	// LoggerFunc optionally retrieves DebugLogger from a context.Context.
	LoggerFunc func(context.Context) DebugLogger

	// FilterRule optionally configures filter rule to allow or deny log messages
	// in some packages or files.
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
	// User can also set the environment variable "EZDBG_FILTER_RULE"
	// to configure it at runtime, if available, the environment variable
	// is used when this value is empty.
	FilterRule string

	filter *logfilter.FileNameFilter
}

func (p *Cfg) getLogger(ctx context.Context) DebugLogger {
	if ctx != nil && p.LoggerFunc != nil {
		if lg := p.LoggerFunc(ctx); lg != nil {
			return lg
		}
	}
	return stdLogger{}
}

// DebugLogger is an interface which log an message at DEBUG level.
// It's implemented by *logrus.Logger, *logrus.Entry, *zap.SugaredLogger,
// and many other logging packages.
type DebugLogger interface {
	Debugf(format string, args ...any)
}

// PrintFunc is a function to print the given arguments in format to somewhere.
// It implements the interface `ErrDebugLogger`.
type PrintFunc func(format string, args ...any)

func (f PrintFunc) Debugf(format string, args ...any) { f(format, args...) }

type stdLogger struct{}

const (
	stdLogDepth    = 2
	stdDebugPrefix = "[DEBUG] "
	stdInfoPrefix  = "[INFO] "
	stdWarnPrefix  = "[WARN] "
)

func (stdLogger) Debugf(format string, args ...any) {
	log.Default().Output(stdLogDepth, fmt.Sprintf(stdDebugPrefix+format, args...))
}

func (stdLogger) Infof(format string, args ...any) {
	log.Default().Output(stdLogDepth, fmt.Sprintf(stdInfoPrefix+format, args...))
}

func (stdLogger) Warnf(format string, args ...any) {
	log.Default().Output(stdLogDepth, fmt.Sprintf(stdWarnPrefix+format, args...))
}
