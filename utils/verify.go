package utils

import (
	"errors"
	"github.com/fishjam/go-library/flog"
	"log"
	"reflect"
)

type CheckErrorAction int

const (
	ACTION_FATAL_QUIT CheckErrorAction = iota
	ACTION_LOG_ERROR
)

const (
	_SKIP_LEVEL = 3
)

//Notice:
//  1. when dev, set to ACTION_FATAL_QUIT, so can check error quickly,
//     then can add error logical for the place that once thought could not go wrong
//  2. when released, set to ACTION_LOG_ERROR, so just log error

var verifyAction = ACTION_FATAL_QUIT

// skip 表示跳过几个调用堆栈, 获取真正有意义的代码调用位置
func checkAndHandleError(err error, msg string, action CheckErrorAction, skip int) {
	if err != nil {
		fileName, lineNo, funName := flog.GetCallStackInfo(skip)

		switch action {
		case ACTION_FATAL_QUIT:
			flog.Warnf("%s:%d (%s) FAIL(%s), msg=%s\n", fileName, lineNo, funName, reflect.TypeOf(err).String(), msg)
		case ACTION_LOG_ERROR:
			log.Panicf("%s:%d (%s) FAIL(%s), msg=%s\n", fileName, lineNo, funName, reflect.TypeOf(err).String(), msg)
			//flog.Infof("error at: %s:%d, msg=%s, err=%s", fileName, lineNo, msg, err)
		}
	}
}

func Verify(err error) error {
	if err != nil {
		checkAndHandleError(err, err.Error(), verifyAction, _SKIP_LEVEL)
	}
	return err
}

func VerifyExcept1(err error, ex1 error) error {
	if err != nil && !errors.Is(ex1, err) {
		checkAndHandleError(err, err.Error(), verifyAction, _SKIP_LEVEL)
	}
	return err
}

func VerifyWithResult[T any](result T, err error) T {
	if err != nil {
		checkAndHandleError(err, err.Error(), verifyAction, _SKIP_LEVEL)
	}
	return result
}

func VerifyWithResultEx[T any](result T, err error) (T, error) {
	if err != nil {
		checkAndHandleError(err, err.Error(), verifyAction, _SKIP_LEVEL)
	}
	return result, err
}
