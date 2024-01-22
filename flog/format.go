package flog

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path"
	"strings"
	"time"
)

// 输出能够让 LogViewer 方便分析的日志
type FLogFormatter struct {
	FullPath        bool
	TimestampFormat string
}

func NewFLogFormatter() *FLogFormatter {
	return &FLogFormatter{
		FullPath:        false,
		TimestampFormat: "2006-01-02T15:04:05.000",
	}
}

func (f *FLogFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	//Entry 里面有所有的日志记录, 包括所有字段集合 等

	timestampFormat := f.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = time.DateTime
	}
	strTime := entry.Time.Format(f.TimestampFormat)
	strLevel := strings.ToUpper(entry.Level.String())

	b := &bytes.Buffer{}
	goId := entry.Data["goid"]
	if goId == nil {
		goId = -1 // 表示没有通过 hook 设置 goid
	}
	traceId := entry.Data["traceId"]
	if traceId == nil {
		traceId = "none"
	}
	callPos := entry.Data["callPos"].(callPosInfo)
	filePath := callPos.file
	if !f.FullPath {
		filePath = path.Base(filePath)
	}
	//2023-11-21T10:24:07.305 [ runtime/proc.go:267 ] [17740][1][INFO] Initialize Logger
	b.WriteString(fmt.Sprintf("%s [ %s:%d ][%d][%d][%s][%s] %s\n",
		strTime, filePath, callPos.line,
		os.Getpid(), goId, strLevel,
		traceId,
		entry.Message))

	//b.WriteString(fmt.Sprintf("%s|%s|%d|", strTime, strLevel, goId))
	//if entry.HasCaller() {
	//	b.WriteString(fmt.Sprintf(" %s:%d ", entry.Caller.File, entry.Caller.Line))
	//}
	//b.WriteString(fmt.Sprintf("%s \n", entry.Message))

	return b.Bytes(), nil
}
