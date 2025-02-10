package zlog

import (
	"bytes"
	"log/slog"
	"testing"

	slogconsolehandler "github.com/jxskiss/slog-console-handler"
	"github.com/stretchr/testify/assert"
)

func TestPrintfFunctions(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	tester := slogconsolehandler.New(buf,
		&slogconsolehandler.HandlerOptions{AddSource: true, Level: slog.LevelDebug})
	SetDefault(slog.New(tester))

	Debugf("Debugf a= %v, b= %v", 123, "abc")
	Infof("Infof a= %v, b= %v", 123, "abc")
	Warnf("Warnf a= %v, b= %v", 123, "abc")
	Errorf("Errorf a= %v, b= %v", 123, "abc")
	Print("Debug: Print a= 123, b= abc")
	Printf("[Warn] Printf a= %v, b= %v", 123, "abc")
	Println("Error: Println a=", 123, "b=", "abc")

	got := buf.String()
	assert.Regexp(t, `DEBUG\s+ Debugf a= 123, b= abc.+source=.*zlog/printf_test.go:\d+`, got)
	assert.Regexp(t, `INFO\s+ Infof a= 123, b= abc.+source=.*zlog/printf_test.go:\d+`, got)
	assert.Regexp(t, `WARN\s+ Warnf a= 123, b= abc.+source=.*zlog/printf_test.go:\d+`, got)
	assert.Regexp(t, `ERROR\s+ Errorf a= 123, b= abc.+source=.*zlog/printf_test.go:\d+`, got)
	assert.Regexp(t, `DEBUG\s+ Debug: Print a= 123, b= abc.+source=.*zlog/printf_test.go:\d+`, got)
	assert.Regexp(t, `WARN\s+ \[Warn\] Printf a= 123, b= abc.+source=.*zlog/printf_test.go:\d+`, got)
	assert.Regexp(t, `ERROR\s+ Error: Println a= 123 b= abc.+source=.*zlog/printf_test.go:\d+`, got)
}
