package zlog

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestAddFields(t *testing.T) {
	ctx := context.Background()
	ctx = AddFields(ctx,
		zap.String("a", "aValue"),
		zap.Int("b", 1234),
	)
	fs := GetFields(ctx)
	require.Len(t, fs, 2)
	assert.Equal(t, fs[0].Key, "a")
	assert.Equal(t, fs[1].Key, "b")

	ctx = AddFields(ctx,
		zap.Duration("b", time.Minute), // override "b":"1234"
	)
	fs = GetFields(ctx)
	require.Len(t, fs, 2)
	assert.Equal(t, fs[1].Key, "b")
	assert.Equal(t, fs[1].Integer, int64(time.Minute))
}

func TestGetLogger(t *testing.T) {

	ctx := context.Background()
	helperReplace := func() (*bytes.Buffer, func()) {
		var buf = &bytes.Buffer{}
		l, p, err := NewWithOutput(&Config{Development: false, Level: "info"}, zapcore.AddSync(buf))
		if err != nil {
			panic(err)
		}
		resetFunc := ReplaceGlobals(l, p)
		return buf, resetFunc
	}

	t.Run("with logger", func(t *testing.T) {
		buf, rf := helperReplace()
		defer rf()

		ctx2 := WithLogger(ctx, L().With(zap.String("a", "aValue")))
		GetLogger(ctx2).Info("test message")
		got := buf.String()
		assert.Contains(t, got, "test message")
		assert.Contains(t, got, `"a":"aValue"`)
	})

	t.Run("with sugared logger", func(t *testing.T) {
		buf, rf := helperReplace()
		defer rf()

		ctx2 := WithLogger(ctx, S().With(zap.Int64("a", 12345), "b", "bValue"))
		GetLogger(ctx2).Info("test message")
		got := buf.String()
		assert.Contains(t, got, "test message")
		assert.Contains(t, got, `"a":12345`)
		assert.Contains(t, got, `"b":"bValue"`)
	})

	t.Run("no logger / with fields", func(t *testing.T) {
		buf, rf := helperReplace()
		defer rf()

		ctx2 := AddFields(ctx, zap.String("a", "aValue"))
		GetLogger(ctx2).Info("test message")
		got := buf.String()
		assert.Contains(t, got, "test message")
		assert.Contains(t, got, `"a":"aValue"`)
	})

	t.Run("no logger / with fields / with extra", func(t *testing.T) {
		buf, rf := helperReplace()
		defer rf()

		ctx2 := AddFields(ctx, zap.String("a", "aValue"))
		GetLogger(ctx2, zap.Int("b", 12345)).Info("test message", zap.Int32("c", 12345))
		got := buf.String()
		assert.Contains(t, got, "test message")
		assert.Contains(t, got, `"a":"aValue"`)
		assert.Contains(t, got, `"b":12345`)
		assert.Contains(t, got, `"c":12345`)
	})

	t.Run("no logger / no fields / no extra", func(t *testing.T) {
		buf, rf := helperReplace()
		defer rf()

		GetLogger(ctx).Info("test message")
		got := buf.String()
		assert.Contains(t, got, "test message")
		assert.NotContains(t, got, `"a":"aValue"`)
		assert.NotContains(t, got, `"b":12345`)
	})
}

func TestGetSugaredLogger(t *testing.T) {
	ctx := context.Background()
	helperReplace := func() (*bytes.Buffer, func()) {
		var buf = &bytes.Buffer{}
		l, p, err := NewWithOutput(&Config{Development: false, Level: "info"}, zapcore.AddSync(buf))
		if err != nil {
			panic(err)
		}
		resetFunc := ReplaceGlobals(l, p)
		return buf, resetFunc
	}

	t.Run("with sugared logger", func(t *testing.T) {
		buf, rf := helperReplace()
		defer rf()

		ctx2 := WithLogger(ctx, S().With("a", "aValue"))
		GetSugaredLogger(ctx2).Info("test message")
		got := buf.String()
		assert.Contains(t, got, "test message")
		assert.Contains(t, got, `"a":"aValue"`)
	})

	t.Run("with extra", func(t *testing.T) {
		buf, rf := helperReplace()
		defer rf()

		ctx2 := WithLogger(ctx, L().With(zap.String("a", "aValue")))
		GetSugaredLogger(ctx2, zap.Int32("a", 12345), zap.String("b", "bValue")).
			Infof("test message")
		got := buf.String()
		assert.Contains(t, got, "test message")
		assert.Contains(t, got, `"a":"aValue"`)
		assert.Contains(t, got, `"a":12345`)
		assert.Contains(t, got, `"b":"bValue"`)
	})
}
