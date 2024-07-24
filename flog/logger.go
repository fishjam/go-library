package flog

import (
	"log"
)

// ILogger in go-library, just need Warnf, so can used in verify,
// user should call `SetLoggerFactory` to customize the logger
type ILogger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

// LoggerFactory is the factory method for creating logger used for the specified package.
type LoggerFactory func() ILogger

func SetLoggerFactory(factory LoggerFactory) {
	_curLogger = factory()
}

type defaultLogger struct {
}

func (l *defaultLogger) Debugf(format string, args ...any) {
	if l != nil {
		log.Printf("[DEBUG] "+format, args...)
	}
}

func (l *defaultLogger) Infof(format string, args ...any) {
	if l != nil {
		log.Printf("[INFO] "+format, args...)
	}
}

func (l *defaultLogger) Warnf(format string, args ...any) {
	if l != nil {
		log.Printf("[WARN] "+format, args...)
	}
}

func (l *defaultLogger) Errorf(format string, args ...any) {
	if l != nil {
		log.Printf("[ERROR] "+format, args...)
	}
}

var _curLogger ILogger = &defaultLogger{}

func Debugf(format string, args ...any) {
	_curLogger.Debugf(format, args...)
}

func Infof(format string, args ...any) {
	_curLogger.Infof(format, args...)
}

func Warnf(format string, args ...any) {
	_curLogger.Warnf(format, args...)
}

func Errorf(format string, args ...any) {
	_curLogger.Errorf(format, args...)
}
