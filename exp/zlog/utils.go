package zlog

import (
	"log"
	"reflect"
	"unsafe"

	"go.uber.org/zap"
)

// log_std links to log.std to get correct caller depth for both
// with and without calling RedirectStdLog.
//go:linkname log_std log.std
var log_std *log.Logger

var zapLoggerNameOffset uintptr

func init() {
	loggerTyp := reflect.TypeOf(zap.Logger{})
	nameField, ok := loggerTyp.FieldByName("name")
	if !ok {
		panic("bug: zap.Logger name field not found")
	}
	zapLoggerNameOffset = nameField.Offset
}

func getLoggerName(l *zap.Logger) string {
	return *(*string)(unsafe.Pointer(uintptr(unsafe.Pointer(l)) + zapLoggerNameOffset))
}
