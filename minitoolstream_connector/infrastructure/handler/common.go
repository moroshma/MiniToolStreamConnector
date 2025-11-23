package handler

import "log"

// Logger defines the logging interface
type Logger interface {
	Printf(format string, v ...interface{})
}

// defaultLogger is a default logger implementation
type defaultLogger struct{}

func (l *defaultLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}
