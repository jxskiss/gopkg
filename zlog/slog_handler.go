//go:build go1.21

package zlog

import (
	"context"
	"log/slog"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ slog.Handler = (*slogHandlerImpl)(nil)

// For Go1.21+, we also replace the default logger in slog package.
func init() {
	replaceSlogDefault = func(l *zap.Logger, disableCaller bool) func() {
		old := slog.Default()
		slog.SetDefault(NewSlogLogger(func(opts *SlogOptions) {
			opts.Logger = l
			opts.DisableCaller = disableCaller
		}))
		return func() { slog.SetDefault(old) }
	}
}

// SetSlogDefault replaces the slog package's default logger.
// Compared to [slog.SetDefault], it does not set the log package's default
// logger to l's handler, instead it sets it's output to an underlying
// zap.Logger returned by L().
func SetSlogDefault(l *slog.Logger) {
	slog.SetDefault(l)
	replaceLogDefault(L().Logger)
}

// NewSlogLogger creates a new slog.Logger.
func NewSlogLogger(options ...func(*SlogOptions)) *slog.Logger {
	opts := newSlogOptions(options)
	impl := &slogHandlerImpl{
		opts: opts,
		l:    opts.Logger,
		name: opts.Logger.Name(),
	}
	return slog.New(impl)
}

func newSlogOptions(options []func(*SlogOptions)) *SlogOptions {
	opts := &SlogOptions{}
	for _, f := range options {
		f(opts)
	}
	// Set defaults.
	if opts.Logger == nil {
		opts.Logger = L().Logger
	}
	return opts
}

// SlogOptions customizes the behavior of slog logger created by NewSlogLogger.
type SlogOptions struct {
	// Logger optionally configures a zap.Logger to use instead of
	// the default logger.
	Logger *zap.Logger

	// DisableCaller stops annotating logs with calling function's file
	// name and line number. By default, all logs are annotated.
	DisableCaller bool

	// ReplaceAttr is called to rewrite an attribute before it is logged.
	// The attribute's value has been resolved (see [slog.Value.Resolve]).
	// If ReplaceAttr returns a zero ReplaceResult, the attribute is discarded.
	//
	// ReplaceAttr may return a single field by ReplaceResult.Field, or return
	// multi fields by ReplaceResult.Multi.
	// User may use this option to check error attribute and convert to a
	// zap.Error field to unify the key for errors, if the error contains
	// stack information, the function can also get it and add the
	// stacktrace to log.
	//
	// Note, for simplicity and better performance, it is different with
	// [slog.HandlerOptions.ReplaceAttr], only attributes directly passed to
	// [slog.Logger.With] and the log methods
	// (such as [slog.Logger.Log], [slog.Logger.Info], etc.)
	// are passed to this function, the builtin attributes and nested attributes
	// in groups are not.
	// The first argument is a list of currently open groups that added by
	// [slog.Logger.WithGroup].
	ReplaceAttr func(groups []string, a slog.Attr) (rr ReplaceResult)
}

// ReplaceResult is a result returned by SlogOptions.ReplaceAttr.
// If there is only one field, set it to Field, else set multi fields to Multi.
// The logger checks both Field and Multi to append to result log.
//
// Note that the field Multi must not contain values whose Type is
// UnknownType or SkipType.
type ReplaceResult struct {
	Field zapcore.Field
	Multi []zapcore.Field
}

func (rr ReplaceResult) hasValidField() bool {
	return (rr.Field.Type != zapcore.UnknownType && rr.Field != zap.Skip()) || len(rr.Multi) > 0
}

type slogHandlerImpl struct {
	opts      *SlogOptions
	l         *zap.Logger
	name      string   // logger name
	allGroups []string // all groups started with WithGroup
	groups    []string // groups that not converted to namespace field
}

func (h *slogHandlerImpl) Enabled(ctx context.Context, level slog.Level) bool {
	zLevel := slogToZapLevel(level)
	ctxFunc := globals.Props.cfg.CtxHandler.ChangeLevel
	if ctx == nil || ctxFunc == nil {
		return h.l.Core().Enabled(zLevel)
	}
	if ctxLevel := ctxFunc(ctx); ctxLevel != nil {
		return ctxLevel.Enabled(zLevel)
	}
	return h.l.Core().Enabled(zLevel)
}

func (h *slogHandlerImpl) Handle(ctx context.Context, record slog.Record) error {
	var ctxResult CtxResult
	ctxFunc := globals.Props.cfg.CtxHandler.WithCtx
	if ctx != nil && ctxFunc != nil {
		ctxResult = ctxFunc(ctx)
	}
	core := h.l.Core()
	if ctxResult.Level != nil {
		core = changeLevel(*ctxResult.Level)(core)
	}

	ent := zapcore.Entry{
		Level:      slogToZapLevel(record.Level),
		Time:       record.Time,
		LoggerName: h.name,
		Message:    record.Message,
	}
	ce := core.Check(ent, nil)
	if ce == nil {
		return nil
	}

	// Add caller information.
	if record.PC > 0 && !h.opts.DisableCaller {
		fs := runtime.CallersFrames([]uintptr{record.PC})
		frame, _ := fs.Next()
		ce.Caller = zapcore.EntryCaller{
			Defined:  true,
			PC:       record.PC, // NOTE: don's use frame.PC here, it's different with record.PC
			File:     frame.File,
			Line:     frame.Line,
			Function: frame.Function,
		}
	}

	var addedNamespace bool
	fields := ctxResult.Fields
	if record.NumAttrs() > 0 {
		guessCap := len(fields) + record.NumAttrs() + len(h.groups) + 2
		fields = make([]zap.Field, 0, guessCap)
		fields = append(fields, ctxResult.Fields...)
	}
	record.Attrs(func(attr slog.Attr) bool {
		rr := h.convertAttrToField(attr)
		if !addedNamespace && len(h.groups) > 0 && rr.hasValidField() {
			// Namespace are added only if at least one field is present
			// to avoid creating empty groups.
			fields = h.appendGroups(fields)
			addedNamespace = true
		}
		if rr.Field.Type != zapcore.UnknownType {
			fields = append(fields, rr.Field)
		}
		if len(rr.Multi) > 0 {
			fields = append(fields, rr.Multi...)
		}
		return true
	})

	ce.Write(fields...)
	return nil
}

func (h *slogHandlerImpl) WithAttrs(attrs []slog.Attr) slog.Handler {
	guessCap := len(attrs) + len(h.groups) + 2
	fields := make([]zapcore.Field, 0, guessCap)
	var addedNamespace bool
	for _, attr := range attrs {
		rr := h.convertAttrToField(attr)
		if !addedNamespace && len(h.groups) > 0 && rr.hasValidField() {
			// Namespace are added only if at least one field is present
			// to avoid creating empty groups.
			fields = h.appendGroups(fields)
			addedNamespace = true
		}
		if rr.Field.Type != zapcore.UnknownType {
			fields = append(fields, rr.Field)
		}
		if len(rr.Multi) > 0 {
			fields = append(fields, rr.Multi...)
		}
	}
	clone := *h
	clone.l = h.l.With(fields...)
	if addedNamespace {
		clone.groups = nil
	}
	return &clone
}

func (h *slogHandlerImpl) appendGroups(fields []zapcore.Field) []zapcore.Field {
	for _, g := range h.groups {
		fields = append(fields, zap.Namespace(g))
	}
	return fields
}

func (h *slogHandlerImpl) WithGroup(name string) slog.Handler {
	// If the name is empty, WithGroup returns the receiver.
	if name == "" {
		return h
	}

	newGroups := make([]string, len(h.allGroups)+1)
	copy(newGroups, h.allGroups)
	newGroups[len(h.allGroups)] = name

	clone := *h
	clone.allGroups = newGroups
	clone.groups = newGroups[len(h.allGroups)-len(h.groups):]
	return &clone
}

func (h *slogHandlerImpl) convertAttrToField(attr slog.Attr) ReplaceResult {
	// Optionally replace attrs.
	// attr.Value is resolved before calling ReplaceAttr, so the user doesn't have to.
	if attr.Value.Kind() == slog.KindLogValuer {
		attr.Value = attr.Value.Resolve()
	}
	if repl := h.opts.ReplaceAttr; repl != nil {
		return repl(h.allGroups, attr)
	}
	return ReplaceResult{Field: ConvertAttrToField(attr)}
}

func (h *slogHandlerImpl) GetUnderlying() *zap.Logger {
	return h.opts.Logger
}

func slogToZapLevel(l slog.Level) zapcore.Level {
	switch {
	case l >= slog.LevelError:
		return ErrorLevel
	case l >= slog.LevelWarn:
		return WarnLevel
	case l >= slog.LevelInfo:
		return InfoLevel
	case l >= slog.LevelDebug:
		return DebugLevel
	default:
		return TraceLevel
	}
}

// groupObject holds all the Attrs saved in a slog.GroupValue.
type groupObject []slog.Attr

func (gs groupObject) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	for _, attr := range gs {
		ConvertAttrToField(attr).AddTo(enc)
	}
	return nil
}

// ConvertAttrToField converts a slog.Attr to a zap.Field.
func ConvertAttrToField(attr slog.Attr) zapcore.Field {
	if attr.Equal(slog.Attr{}) {
		// Ignore empty attrs.
		return zap.Skip()
	}

	switch attr.Value.Kind() {
	case slog.KindBool:
		return zap.Bool(attr.Key, attr.Value.Bool())
	case slog.KindDuration:
		return zap.Duration(attr.Key, attr.Value.Duration())
	case slog.KindFloat64:
		return zap.Float64(attr.Key, attr.Value.Float64())
	case slog.KindInt64:
		return zap.Int64(attr.Key, attr.Value.Int64())
	case slog.KindString:
		return zap.String(attr.Key, attr.Value.String())
	case slog.KindTime:
		return zap.Time(attr.Key, attr.Value.Time())
	case slog.KindUint64:
		return zap.Uint64(attr.Key, attr.Value.Uint64())
	case slog.KindGroup:
		grpAttrs := attr.Value.Group()
		if len(grpAttrs) == 0 {
			return zap.Skip()
		}
		if attr.Key == "" {
			// Inlines recursively.
			return zap.Inline(groupObject(grpAttrs))
		}
		return zap.Object(attr.Key, groupObject(grpAttrs))
	case slog.KindLogValuer:
		return ConvertAttrToField(slog.Attr{
			Key:   attr.Key,
			Value: attr.Value.Resolve(),
		})
	default:
		return zap.Any(attr.Key, attr.Value.Any())
	}
}
