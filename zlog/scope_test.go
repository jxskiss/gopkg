package zlog

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"

	slogconsolehandler "github.com/jxskiss/slog-console-handler"
	"github.com/stretchr/testify/assert"

	"github.com/jxskiss/gopkg/v2/zlog/internal/scopepkg"
)

func TestScope(t *testing.T) {
	t.Run("RegisterScope", func(t *testing.T) {
		resetScopes()
		s := RegisterScope("test", "test description")
		assert.Equal(t, "test", s.Name())
		assert.Equal(t, "test description", s.Description())
		assert.Equal(t, slog.LevelInfo, s.level.Level())
	})

	t.Run("FindScope", func(t *testing.T) {
		resetScopes()
		_ = RegisterScope("test", "test description")
		s := FindScope("test")
		assert.Equal(t, "test", s.Name())
		assert.Equal(t, "test description", s.Description())
		assert.Nil(t, FindScope("not_exist"))
	})

	t.Run("Scopes", func(t *testing.T) {
		resetScopes()
		_ = RegisterScope("test", "test description")
		s := Scopes()
		assert.Equal(t, 1, len(s))
		assert.Equal(t, "test", s["test"].Name())
	})

	t.Run("change level", func(t *testing.T) {
		resetScopes()
		s := RegisterScope("test", "test description")
		assert.Equal(t, slog.LevelInfo, s.GetLevel())
		s.SetLevel(slog.LevelDebug)
		assert.Equal(t, slog.LevelDebug, s.GetLevel())
	})
}

func TestScope_loggers(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	tester := slogconsolehandler.New(buf,
		&slogconsolehandler.HandlerOptions{AddSource: true, Level: slog.LevelDebug})
	SetDefault(slog.New(NewHandler(tester, nil)))

	t.Run("Logger", func(t *testing.T) {
		resetScopes()
		buf.Reset()
		log1 := RegisterScope("log1", "log1 description")

		logger1 := log1.Logger()
		logger2 := log1.Logger().With("k1", 123).With("k2", "abc")
		logger1.Info("test logger1", "k1", 123, "k2", "abc")
		logger2.Info("test logger2")
		got := buf.String()
		assert.Regexp(t, `INFO\s+ test logger1.+logger= log1.+k1= 123.+k2= abc.+source=.*zlog/scope_test.go:\d+`, got)
		assert.Regexp(t, `INFO\s+ test logger2.+logger= log1.+k1= 123.+k2= abc.+source=.*zlog/scope_test.go:\d+`, got)
	})

	t.Run("WithError", func(t *testing.T) {
		resetScopes()
		buf.Reset()

		ctx := context.Background()
		ctx = PrependAttrs(ctx, "prepend1", "arg1")
		ctx = AppendAttrs(ctx, "append1", "arg1", slog.String("append2", "arg2"))

		log1 := RegisterScope("log1", "log1 description")
		logger1 := log1.WithError(ctx, nil, "k1", 123).With("k2", "abc")
		logger2 := log1.WithError(ctx, errors.New("logger2Error"), "k1", 123).With("k2", "abc")
		logger1.InfoContext(ctx, "test logger1")
		logger2.InfoContext(ctx, "test logger2")
		got := buf.String()
		assert.Regexp(t, `INFO\s+ test logger1.+logger= log1.+prepend1= arg1.+k1= 123.+k2= abc.+append1= arg1.+append2= arg2.+source=.*zlog/scope_test.go:\d+`, got)
		assert.Regexp(t, `INFO\s+ test logger2.+logger= log1.+prepend1= arg1.+k1= 123.+k2= abc.+append1= arg1.+append2= arg2.+source=.*zlog/scope_test.go:\d+`, got)
		assert.Regexp(t, `INFO\s+ test logger2.+error= logger2Error`, got)
	})

	t.Run("WithGroup", func(t *testing.T) {
		resetScopes()
		buf.Reset()

		ctx := context.Background()
		ctx = PrependAttrs(ctx, "prepend1", "arg1")
		ctx = AppendAttrs(ctx, "append1", "arg1", slog.String("append2", "arg2"))

		log1 := RegisterScope("log1", "log1 description")
		logger1 := log1.WithGroup(ctx, "group1", "k1", 123).With("k2", "abc")
		logger2 := log1.WithGroup(ctx, "group2").With("k1", 123).With("k2", "abc")

		logger1.Info("test logger1") // ctx passed by FromCtx
		logger2.Info("test logger2") // ctx passed by FromCtx
		got := buf.String()

		// assert logger1's output
		assert.Regexp(t, `INFO\s+ test logger1.+logger= log1.+prepend1= arg1`, got)
		assert.Regexp(t, `INFO\s+ test logger1.+group1\.k1= 123\s+group1\.k2= abc`, got)
		assert.Regexp(t, `INFO\s+ test logger1.+group1\.append1= arg1\s+group1\.append2= arg2`, got)

		// assert logger2's output
		assert.Regexp(t, `INFO\s+ test logger2.+logger= log1.+prepend1= arg1`, got)
		assert.Regexp(t, `INFO\s+ test logger2.+group2\.k1= 123\s+group2\.k2= abc`, got)
		assert.Regexp(t, `INFO\s+ test logger2.+group2\.append1= arg1\s+group2\.append2= arg2`, got)
	})

	t.Run("from context", func(t *testing.T) {
		resetScopes()
		buf.Reset()
		log1 := RegisterScope("log1", "log1 description")

		ctx := context.Background()
		ctx = PrependAttrs(ctx, "prepend1", "arg1")
		ctx = AppendAttrs(ctx, "append1", "arg1", slog.String("append2", "arg2"))
		ctx = NewCtx(ctx, Default().With("ctx1", 123))

		logger1 := log1.With(ctx, "k1", 123).With("k2", "abc")
		logger1.InfoContext(ctx, "test logger1")
		got := buf.String()
		assert.Regexp(t, `INFO\s+ test logger1.+logger= log1.+prepend1= arg1.+k1= 123.+k2= abc.+append1= arg1.+append2= arg2.+source=.*zlog/scope_test.go:\d+`, got)
		assert.Regexp(t, `INFO\s+ test logger1.+ctx1= 123`, got)
	})

	t.Run("loggerName overriding", func(t *testing.T) {
		resetScopes()
		buf.Reset()

		// setup scopepkg logger
		scopepkg.SetupScoperLogger(RegisterScope("scopepkg", "test scope logger"))

		log1 := RegisterScope("log1", "log1 description")

		ctx := context.Background()
		ctx = PrependAttrs(ctx, "prepend1", "arg1")
		ctx = AppendAttrs(ctx, "append1", "arg1", slog.String("append2", "arg2"))
		ctx = NewCtx(ctx, log1.Logger().With("ctx1", 123))

		FromCtx(ctx).InfoContext(ctx, "test FromCtx")
		got1 := buf.String()
		assert.Regexp(t, `INFO\s+ test FromCtx.+prepend1= arg1.+logger= log1.+ctx1= 123.+append1= arg1.+append2= arg2.+source=.*zlog/scope_test.go:\d+`, got1)

		buf.Reset()
		scopepkg.PrintLog(ctx, "k1", 123, "k2", "abc")
		got2 := buf.String()

		assert.Regexp(t, `INFO\s+ test scope logger.+prepend1= arg1.+logger= scopepkg.+ctx1= 123.+k1= 123.+k2= abc.+append1= arg1.+append2= arg2.+source=.*zlog/internal/scopepkg/scope.go:\d+`, got2)
		assert.NotRegexp(t, `logger= log1`, got2)
	})
}

func TestScope_misuseRecursiveCalling(t *testing.T) {
	resetScopes()

	log1 := RegisterScope("log1", "log1 description").Logger()
	slog.SetDefault(log1) // misuse

	ctx := context.Background()

	assert.NotPanics(t, func() {
		_ = log1.Enabled(ctx, slog.LevelError)
		_ = log1.With("k1", "v1")
		_ = log1.WithGroup("group1")
	})

	assert.Panics(t, func() {
		log1.Info("info message")
	})
	assert.Panics(t, func() {
		log1.With("k1", "v1").Info("info message")
	})
	assert.Panics(t, func() {
		log1.WithGroup("group1").Info("info message")
	})
	assert.Panics(t, func() {
		log1.InfoContext(ctx, "info message")
	})
	assert.Panics(t, func() {
		log1.Log(ctx, slog.LevelInfo, "info message", "k1", "v1")
	})
	assert.Panics(t, func() {
		log1.LogAttrs(ctx, slog.LevelInfo, "info message")
	})
	assert.Panics(t, func() {
		_ = log1.Handler().Handle(ctx, slog.Record{})
	})
}

func resetScopes() {
	scopeLock.Lock()
	defer scopeLock.Unlock()
	scopesTable = make(map[string]*Scope)
}
