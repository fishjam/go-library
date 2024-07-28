package flog

import (
	"log"
	"os"
	"path"
)

// ILogger in go-library, just need Warnf, so can used in verify,
// user should call `SetLoggerFactory` to customize the logger
type ILogger interface {
	WarnExWithPosf(fileName string, lineNo int, funName string, format string, args ...any)
}

// LoggerFactory is the factory method for creating logger used for the specified package.
type LoggerFactory func() ILogger

func SetLoggerFactory(factory LoggerFactory) {
	_curLogger = factory()
}

type defaultLogger struct {
}

func (l *defaultLogger) WarnExWithPosf(fileName string, lineNo int, funName string, format string, args ...interface{}){
	if l != nil {
		log.Printf("[ %s:%d ][%d][%d][WARN][%s] "+format,
			append([]interface{}{ path.Base(fileName), lineNo, os.Getpid(), GetGoroutineID(), "none"},
			args...)...)
	}
}

var _curLogger ILogger = &defaultLogger{}

func WarnExWithPosf(fileName string, lineNo int, funName string, format string, args ...any) {
	_curLogger.WarnExWithPosf(fileName, lineNo, funName, format, args...)
}