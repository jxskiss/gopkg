package zlog

import (
	"bytes"
	"log"
	"log/slog"
	"testing"

	slogconsolehandler "github.com/jxskiss/slog-console-handler"
	"github.com/stretchr/testify/assert"
)

func TestRedirectStdLog(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	logger := slog.New(slogconsolehandler.New(buf,
		&slogconsolehandler.HandlerOptions{AddSource: true, Level: slog.LevelDebug}))
	RedirectStdLog(logger, []slog.Attr{
		slog.String("_logger", "redirect_test"),
	})

	log.Println("Debug: message 1")
	log.Printf("[Info] message 2")
	log.Print("Warn: message 3")
	log.Print("[Warning] message 4")
	log.Print("Warning: message 5")
	log.Print("Error: message 6")
	log.Print("fatal: message 7")
	log.Print("panic: message 8")
	log.Print("fatal message 9")
	log.Print("Error message 10")

	got := buf.String()
	assert.Regexp(t, `DEBUG\s+ Debug: message 1.+_logger= redirect_test.+source=.*zlog/redirect_test.go:\d+`, got)
	assert.Regexp(t, `INFO\s+ \[Info\] message 2.+_logger= redirect_test.+source=.*zlog/redirect_test.go:\d+`, got)
	assert.Regexp(t, `WARN\s+ Warn: message 3.+_logger= redirect_test.+source=.*zlog/redirect_test.go:\d+`, got)
	assert.Regexp(t, `WARN\s+ \[Warning\] message 4.+_logger= redirect_test.+source=.*zlog/redirect_test.go:\d+`, got)
	assert.Regexp(t, `WARN\s+ Warning: message 5.+_logger= redirect_test.+source=.*zlog/redirect_test.go:\d+`, got)
	assert.Regexp(t, `ERROR\s+ Error: message 6.+_logger= redirect_test.+source=.*zlog/redirect_test.go:\d+`, got)
	assert.Regexp(t, `ERROR\s+ fatal: message 7.+_logger= redirect_test.+source=.*zlog/redirect_test.go:\d+`, got)
	assert.Regexp(t, `ERROR\s+ panic: message 8.+_logger= redirect_test.+source=.*zlog/redirect_test.go:\d+`, got)
	assert.Regexp(t, `INFO\s+ fatal message 9.+_logger= redirect_test.+source=.*zlog/redirect_test.go:\d+`, got)
	assert.Regexp(t, `INFO\s+ Error message 10.+_logger= redirect_test.+source=.*zlog/redirect_test.go:\d+`, got)
}
