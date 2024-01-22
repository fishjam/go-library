package flog

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"os"
	"path/filepath"
)

var flog = setupLogger()

func setupLogger() *logrus.Logger {
	logRootDir := "logs"
	fjsdkPath := os.Getenv("FJSDK")
	if len(fjsdkPath) > 0 {
		logRootDir = filepath.Join(fjsdkPath, "GoStudy", "logs")
	} else {
		panic("must setup FJSDK env variable")
	}

	err := os.MkdirAll(logRootDir, 0755)
	if err != nil {
		panic(fmt.Sprintf("mkdir %s fail, err=%+v", logRootDir, err))
	}

	//lr := logrus.StandardLogger()
	lr := logrus.New()
	fmt.Printf("")
	rotateLogger := &lumberjack.Logger{
		Filename:   filepath.Join(logRootDir, "go_study.log"),
		MaxSize:    50, // megabytes
		MaxBackups: 3,
		MaxAge:     7,     //days
		Compress:   false, // disabled by default
	}

	lr.SetOutput(io.MultiWriter(os.Stdout, rotateLogger))

	lr.SetReportCaller(true)
	lr.AddHook(NewFLogHook())
	lr.SetFormatter(NewFLogFormatter())

	return lr
}

func Debugf(format string, args ...interface{}) {
	flog.WithContext(nil).Debugf(format, args...)
}

func DebugExf(format string, args ...interface{}) {
	flog.WithContext(nil).Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	flog.WithContext(nil).Infof(format, args...)
}

func InfoExf(ctx context.Context, format string, args ...interface{}) {
	flog.WithContext(ctx).Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	flog.WithContext(nil).Warnf(format, args...)
}

func WarnExf(ctx context.Context, format string, args ...interface{}) {
	flog.WithContext(ctx).Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	flog.WithContext(nil).Errorf(format, args...)
}

func ErrorExf(ctx context.Context, format string, args ...interface{}) {
	flog.WithContext(ctx).Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	flog.Fatalf(format, args...)
}

func FatalExf(ctx context.Context, format string, args ...interface{}) {
	flog.WithContext(ctx).Fatalf(format, args...)
}
