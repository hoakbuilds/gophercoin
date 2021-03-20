package log

import (
	"os"
)

// Logger interface of a Log. Useful for mocks and generic logs
type Logger interface {
	WithError(err error) Logger
	WithDetails(details ...Detail) Logger
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Error(format string, args ...interface{})
}

// NewLogger returns a new Logger
func NewLogger(level Level) Logger {
	return New(Conf{
		Writer: os.Stdout,
		Level:  level,
	})
}

// Fatal sends an alarm message and exits program
func Fatal(l Logger, err error, msg string) {
	l.WithError(err)
	panic(err)
}
