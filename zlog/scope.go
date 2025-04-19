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

func (s *Scope) Logger() *Logger {
	h0 := &wrapDefaultHandler{level: s.level}
	h1 := &Handler{
		level: s.level,
		next:  h0,
	}
	h1.goa = h1.goa.withLoggerName(s.name)
	return slog.New(h1)
}

func (s *Scope) With(ctx context.Context, args ...any) *Logger {
	return s.newScopeLogger(FromCtx(ctx)).With(args...)
}

func (s *Scope) WithError(ctx context.Context, err error, args ...any) *Logger {
	if err == nil {
		return s.With(ctx, args...)
	}
	return s.newScopeLogger(FromCtx(ctx)).With(slog.Any(ErrorKey, err)).With(args...)
}

func (s *Scope) WithGroup(ctx context.Context, group string, args ...any) *Logger {
	if group == "" {
		return s.With(ctx, args...)
	}
	return s.newScopeLogger(FromCtx(ctx)).WithGroup(group).With(args...)
}

func (s *Scope) newScopeLogger(l *Logger) *Logger {
	if h, ok := l.Handler().(*Handler); ok {
		return slog.New(h.withScope(s.name, s.level))
	}
	h := &Handler{
		level: s.level,
		next:  l.Handler(),
	}
	h.goa = h.goa.withLoggerName(s.name)
	return slog.New(h)
}

type wrapDefaultHandler struct {
	level *slog.LevelVar
}

func (h *wrapDefaultHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *wrapDefaultHandler) Handle(ctx context.Context, record slog.Record) error {
	return Default().Handler().Handle(ctx, record)
}

func (h *wrapDefaultHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return Default().Handler().WithAttrs(attrs)
}

func (h *wrapDefaultHandler) WithGroup(name string) slog.Handler {
	return Default().Handler().WithGroup(name)
}
