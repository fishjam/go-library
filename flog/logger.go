package flog

import (
	"log"
	"os"
	"path"
	"runtime"
)

// ILogger in go-library, provide WarnExWithPosf, so can used in verify,
// user should call `SetLoggerFactory` to customize the logger
type ILogger interface {
	Debugf(format string, args ...any)
	Infof(format string, args ...any)
	WarnExWithPosf(fileName string, lineNo int, funName string, format string, args ...any)
}

// LoggerFactory is the factory method for creating logger used for the specified package.
type LoggerFactory func() ILogger

func SetLoggerFactory(factory LoggerFactory) {
	_curLogger = factory()
}

type defaultLogger struct {
	// TODO: define another flog.Level?
	// same as : logrus's Level: Trace(6)>Debug(5)>Info(4)>Warn(3)>Error(2)>Fatal(1)>Panic(0)
	level int
}

func (l *defaultLogger) innerLog(level, format string, args ...any) {
	_, fileName, lineNo, _ := runtime.Caller(3)
	log.Printf("[ %s:%d ][%d][%d][%s][%s] "+format,
		append([]interface{}{path.Base(fileName), lineNo, os.Getpid(), GetGoroutineID(), level, "none"},
			args...)...)
}

func (l *defaultLogger) Debugf(format string, args ...any) {
	if l != nil && l.level >= 5 { // >= Debug(5)
		l.innerLog("Debug", format, args...)
	}
}

func (l *defaultLogger) Infof(format string, args ...any) {
	if l != nil && l.level >= 4 { // >= Info(4)
		l.innerLog("Info", format, args...)
	}
}

func (l *defaultLogger) WarnExWithPosf(fileName string, lineNo int, funName string, format string, args ...interface{}) {
	if l != nil && l.level >= 3 { // >= Warn(3)
		log.Printf("[ %s:%d ][%d][%d][WARN][%s] "+format,
			append([]interface{}{path.Base(fileName), lineNo, os.Getpid(), GetGoroutineID(), "none"},
				args...)...)
	}
}

var _curLogger ILogger = &defaultLogger{
	level: 5, //default is 3(warn)
}

func Debugf(format string, args ...any) {
	_curLogger.Debugf(format, args...)
}

func Infof(format string, args ...any) {
	_curLogger.Infof(format, args...)
}

func WarnExWithPosf(fileName string, lineNo int, funName string, format string, args ...any) {
	_curLogger.WarnExWithPosf(fileName, lineNo, funName, format, args...)
}