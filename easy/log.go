package easy

import "log"

type ErrLogger interface {
	Errorf(format string, args ...interface{})
}

type Printer interface {
	Printf(format string, args ...interface{})
}

func logErr(logger interface{}, format string, args ...interface{}) {
	switch logger := logger.(type) {
	case ErrLogger:
		logger.Errorf(format, args...)
	case Printer:
		logger.Printf(format, args...)
	case func(format string, args ...interface{}):
		logger(format, args...)
	default:
		log.Printf(format, args...)
	}
}
