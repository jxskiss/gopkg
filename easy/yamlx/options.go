package yamlx

// Option customizes the behavior of the extended YAML parser.
type Option struct {
	apply func(*extOptions)
}

type extOptions struct {
	EnableEnv     bool
	EnableInclude bool
	IncludeDirs   []string
	FuncMap       FuncMap
}

func (o *extOptions) apply(opts ...Option) *extOptions {
	for _, opt := range opts {
		opt.apply(o)
	}
	return o
}

// EnableEnv enables reading environment variables.
// By default, it is disabled for security considerations.
func EnableEnv() Option {
	return Option{
		apply: func(options *extOptions) {
			options.EnableEnv = true
		}}
}

// EnableInclude enables including other files.
// By default, it is disabled for security considerations.
func EnableInclude() Option {
	return Option{
		apply: func(options *extOptions) {
			options.EnableInclude = true
		}}
}

// WithIncludeDirs optionally specifies the directories to find include files.
// By default, the current working directory is used to search include files.
func WithIncludeDirs(dirs ...string) Option {
	return Option{
		apply: func(options *extOptions) {
			options.IncludeDirs = dirs
		}}
}

// FuncMap is the type of the map defining the mapping from names to functions.
// Each function must have either a single return value, or two return values of
// which the second is an error.
// In case the second return value evaluates to a non-nil error during execution,
// the execution terminates and the error will be returned.
type FuncMap map[string]any

// WithFuncMap specifies additional functions to use with the "@@fn" directive.
func WithFuncMap(funcMap FuncMap) Option {
	return Option{
		apply: func(options *extOptions) {
			options.FuncMap = funcMap
		}}
}
