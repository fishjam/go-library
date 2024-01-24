package flog

import (
	"log"
)

// ILogger in go-library, just need Warnf, so can used in verify,
// user should call `SetLoggerFactory` to customize the logger
type ILogger interface {
	Warnf(format string, args ...any)
}

// LoggerFactory is the factory method for creating logger used for the specified package.
type LoggerFactory func() ILogger

func SetLoggerFactory(factory LoggerFactory) {
	_curLogger = factory()
}

type defaultLogger struct {
}

func (l *defaultLogger) Warnf(format string, args ...any) {
	log.Printf(format, args...)
}

var _curLogger ILogger = &defaultLogger{}

func Warnf(format string, args ...any) {
	_curLogger.Warnf(format, args...)
}
