package zlog

import (
	"context"
	"log/slog"
	"sync"
)

// Scope constrains logging control to a named scope level.
// It gives users a fine-grained control over output severity.
type Scope struct {
	name        string
	description string
	level       *slog.LevelVar
}

var (
	scopeLock   sync.RWMutex
	scopesTable = make(map[string]*Scope)
)

// RegisterScope registers a new logging scope. If the same name is used multiple times
// for a single process, the same Scope struct is returned.
func RegisterScope(name, description string) *Scope {
	scopeLock.Lock()
	defer scopeLock.Unlock()

	s, ok := scopesTable[name]
	if !ok {
		s = &Scope{
			name:        name,
			description: description,
			level:       &slog.LevelVar{},
		}
		scopesTable[name] = s
	}
	return s
}

// FindScope returns a previously registered scope,
// or nil if the named scope wasn't previously registered.
func FindScope(name string) *Scope {
	scopeLock.RLock()
	defer scopeLock.RUnlock()
	return scopesTable[name]
}

// Scopes returns a snapshot of the currently defined set of scopes.
func Scopes() map[string]*Scope {
	scopeLock.RLock()
	defer scopeLock.RUnlock()

	out := make(map[string]*Scope)
	for k, v := range scopesTable {
		out[k] = v
	}
	return out
}

// Name returns this scope's name.
func (s *Scope) Name() string { return s.name }

// Description returns this scope's description.
func (s *Scope) Description() string { return s.description }

// GetLevel returns the output level associated with the scope.
func (s *Scope) GetLevel() slog.Level { return s.level.Level() }

// SetLevel changes the output level associated with the scope.
func (s *Scope) SetLevel(level slog.Level) { s.level.Set(level) }

// Logger returns a scoped logger wrapping slog's default logger.
// For most use-case, methods [Scope.With], [Scope.WithError], [Scope.WithGroup]
// are better choices for context-aware logging.
//
// IMPORTANT NOTE:
// the returned logger must not be used as slog's default logger by
// calling slog.SetDefault, which leads to infinite recursive calling.
func (s *Scope) Logger() *Logger {
	h0 := &proxyDefaultHandler{}
	h1 := &Handler{
		next:  h0,
		scope: s,
	}
	return slog.New(h1)
}

func (s *Scope) With(ctx context.Context, args ...any) *Logger {
	h0 := fromCtxHandler(ctx)
	h1 := h0.withScope(s).withArgs(args)
	return slog.New(h1)
}

func (s *Scope) WithError(ctx context.Context, err error, args ...any) *Logger {
	if err == nil {
		return s.With(ctx, args...)
	}
	h0 := fromCtxHandler(ctx)
	h1 := h0.withScope(s).
		WithAttrs([]slog.Attr{slog.Any(ErrorKey, err)}).(*Handler).
		withArgs(args)
	return slog.New(h1)
}

func (s *Scope) WithGroup(ctx context.Context, group string, args ...any) *Logger {
	if group == "" {
		return s.With(ctx, args...)
	}
	h0 := fromCtxHandler(ctx)
	h1 := h0.withScope(s).WithGroup(group).(*Handler).withArgs(args)
	return slog.New(h1)
}

type proxyDefaultHandler struct{}

func (*proxyDefaultHandler) Handle(ctx context.Context, record slog.Record) error {
	// In case of misuse, setting this handler to slog.Default() leads to
	// infinite recursive calling, which exhausts all CPU and memory resource.
	// We use ctx marker to detect recursive calling.
	type ctxMarker struct{}
	marker, _ := ctx.Value(ctxMarker{}).(int)
	if marker > 0 {
		panic("bug: zlog scope logger must not be used as slog's default logger")
	}

	// set or update the ctx marker
	ctx = context.WithValue(ctx, ctxMarker{}, marker+1)

	return Default().Handler().Handle(ctx, record)
}

func (*proxyDefaultHandler) Enabled(_ context.Context, _ slog.Level) bool {
	panic("unreachable")
}

func (*proxyDefaultHandler) WithAttrs(_ []slog.Attr) slog.Handler {
	panic("unreachable")
}

func (*proxyDefaultHandler) WithGroup(_ string) slog.Handler {
	panic("unreachable")
}
