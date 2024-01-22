package utils

import (
	"errors"
	"fishjam.com/go-library/flog"
	"fmt"
	"log"
	"reflect"
	"runtime"
)

type CheckErrorAction int

const (
	ACTION_FATAL_QUIT CheckErrorAction = iota
	ACTION_LOG_ERROR
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
			fmt.Printf("%s:%d: (%s) FAIL(%s), msg=%s\n", fileName, lineNo, funName, reflect.TypeOf(err).String(), msg)
			log.Fatalf("") //"error at: %s:%d, msg=%s, err=%s", fileName, lineNo, msg, err)
		case ACTION_LOG_ERROR:
			fmt.Printf("%s:%d: (%s) FAIL(%s), msg=%s\n", fileName, lineNo, funName, reflect.TypeOf(err).String(), msg)
			//flog.Infof("error at: %s:%d, msg=%s, err=%s", fileName, lineNo, msg, err)
		}
	}
}

func CheckAndFatalIfError(err error, msg string) {
	checkAndHandleError(err, msg, ACTION_FATAL_QUIT, 1)
}

func Verify(err error) error {
	if err != nil {
		checkAndHandleError(err, err.Error(), verifyAction, 3)
	}
	return err
}

func VerifyExcept1(err error, ex1 error) error {
	if err != nil && !errors.Is(ex1, err) {
		checkAndHandleError(err, err.Error(), verifyAction, 3)
	}
	return err
}

func VerifyWithResult(result interface{}, err error) interface{} {
	if err != nil {
		checkAndHandleError(err, err.Error(), verifyAction, 3)
	}
	return result
}

func VerifyWithResultEx(result interface{}, err error) (interface{}, error) {
	if err != nil {
		checkAndHandleError(err, err.Error(), verifyAction, 3)
	}
	return result, err
}

// TODO: 一个简单的 assert ?(可以获取到调用处的函数名等信息), 写个 test 进行测试
func Assert(cond bool) {
	if !cond {
		_, f, l, _ := runtime.Caller(1)
		panic(fmt.Sprintf("%s:%d: assertion failed", f, l))
	}
}
