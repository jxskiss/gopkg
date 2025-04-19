package zlog

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCtxAndFromCtx(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	ctx := NewCtx(nil, logger)

	got1 := FromCtx(context.TODO())
	got2 := FromCtx(context.Background())
	got3 := FromCtx(ctx)
	assert.True(t, got1 != logger)
	assert.True(t, got2 != logger && got1 == got2)
	assert.True(t, got3 == logger)
}
