package flog

import (
	"fishjam.com/go-library/consts"
	"github.com/sirupsen/logrus"
	"runtime"
)

type FLogHook struct {
}

func NewFLogHook() *FLogHook {
	return &FLogHook{}
}

func (hook *FLogHook) Levels() []logrus.Level {
	return logrus.AllLevels
}

type callPosInfo struct {
	file string
	line int
}

func (hook *FLogHook) Fire(entry *logrus.Entry) error {
	// 在日志里面加入 goid
	//entry.Context = context.WithValue(entry.Context, "goid", utils.GetGoroutineID())
	entry.Data["goid"] = GetGoroutineID()

	moreSkip := 0
	if entry.Context != nil {
		if traceId, ok := entry.Context.Value(consts.TRACE_ID).(string); ok {
			//moreSkip = -1
			entry.Data["traceId"] = traceId
		}
	}

	_, file, line, ok := runtime.Caller(8 + moreSkip)
	if ok {
		entry.Data["callPos"] = callPosInfo{
			file: file,
			line: line,
		}
	}

	//pcs := make([]uintptr, 10)
	//depth := runtime.Callers(1, pcs)
	//frames := runtime.CallersFrames(pcs[:depth])
	//frame, _ := frames.Next()
	//entry.Caller = frame

	return nil
}
