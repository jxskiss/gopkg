package lumberjack_writer

import (
	"errors"
	"os"

	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/jxskiss/gopkg/v2/zlog"
)

func NewLumberjackWriter(fc *zlog.FileConfig) (zapcore.WriteSyncer, func(), error) {
	if st, err := os.Stat(fc.Filename); err == nil {
		if st.IsDir() {
			return nil, nil, errors.New("cannot use directory as log filename")
		}
	}
	out := &lumberjack.Logger{
		Filename:   fc.Filename,
		MaxSize:    fc.MaxSize,
		MaxAge:     fc.MaxDays,
		MaxBackups: fc.MaxBackups,
		LocalTime:  true,
		Compress:   fc.Compress,
	}
	writer := zapcore.AddSync(out)
	closer := func() {
		_ = out.Close()
	}
	return writer, closer, nil
}
