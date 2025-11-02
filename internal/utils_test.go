package internal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDefaultErrorLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	slog.SetDefault(slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{
		AddSource: true,
	})))

	ctx := context.Background()
	DefaultLoggerError(ctx, errors.New("test error 1"), "test message 1")
	helperCallDefaultErrorLooger()

	// Check if the log message contains the error and message
	logs := buf.String()
	require.Regexp(t, `level=ERROR source=.+/internal/utils_test\.go:\d+ msg="test message 1" error="test error 1"`, logs)
	require.Regexp(t, `level=ERROR source=.+/internal/utils_test\.go:\d+ msg="test message 2" error="test error 2"`, logs)
}

func helperCallDefaultErrorLooger() {
	fmt.Println("running helperCallDefaultErrorLooger")
	DefaultLoggerError(context.Background(), errors.New("test error 2"), "test message 2")
}
