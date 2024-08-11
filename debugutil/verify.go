package debugutil

import (
	"errors"
	"fmt"
	"github.com/fishjam/go-library/flog"
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

type Config struct {
	MoreSkip 			int
	Message     		string
	IgnoreExceptions 	[]error
}

//Notice:
//  1. when dev, set to ACTION_FATAL_QUIT, so can check error quickly,
//     then can add error logical for the place that once thought could not go wrong
//  2. when released, set to ACTION_LOG_ERROR, so just log when there is error
//  TODO: refactor with go build tag?

var verifyAction = ACTION_LOG_ERROR

// skip 表示跳过几个调用堆栈, 获取真正有意义的代码调用位置
func checkAndHandleError(err error, msg string, action CheckErrorAction, skip int) {
	if err != nil {
		fileName, lineNo, funName := flog.GetCallStackInfo(skip)
		switch action {
		case ACTION_LOG_ERROR:
			flog.WarnExWithPosf(fileName, lineNo, funName, "verify fail: err=%s(%s), msg=%q",
				reflect.TypeOf(err).String(), err.Error(), msg)
		case ACTION_FATAL_QUIT:
			newMsg := fmt.Sprintf("%s:%d (%s) FAIL(%s), msg=%q\n",
				fileName, lineNo, funName, reflect.TypeOf(err).String(), msg)
			panic(newMsg)
		}
	}
}

func CheckAndFatalIfError(err error, msg string) {
	checkAndHandleError(err, msg, ACTION_FATAL_QUIT, _SKIP_LEVEL)
}

func VerifyWithConfig(err error, config *Config) error{
	if err != nil {
		ignore := false
		moreSkip := 0
		msg := err.Error()

		if config != nil {
			for _, ignoreExc := range config.IgnoreExceptions {
				if errors.Is(err, ignoreExc) {
					ignore = true
					break
				}
			}
			if config.Message != "" {
				msg = config.Message
			}
			moreSkip = config.MoreSkip
		}

		if !ignore {
			checkAndHandleError(err, msg, verifyAction, _SKIP_LEVEL + moreSkip)
		}
	}
	return err
}

func Verify(err error) error {
	return VerifyWithConfig(err, nil)
	//if err != nil {
	//	checkAndHandleError(err, err.Error(), verifyAction, _SKIP_LEVEL)
	//}
	//return err
}

//func VerifyMoreSkip(err error, moreSkip int) error {
//	if err != nil {
//		checkAndHandleError(err, err.Error(), verifyAction, _SKIP_LEVEL + moreSkip)
//	}
//	return err
//}

func VerifyWithMessage(err error, msg string) error {
	if err != nil {
		checkAndHandleError(err, msg, verifyAction, _SKIP_LEVEL)
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

// two result without error
func VerifyWithTwoResult[R1 any, R2 any](r1 R1, r2 R2, err error) (R1, R2) {
	if err != nil {
		checkAndHandleError(err, err.Error(), verifyAction, _SKIP_LEVEL)
	}
	return r1, r2
}


// two result with error
func VerifyWithTwoResultEx[R1 any, R2 any](r1 R1, r2 R2, err error) (R1, R2, error) {
	if err != nil {
		checkAndHandleError(err, err.Error(), verifyAction, _SKIP_LEVEL)
	}
	return r1, r2, err
}

func Assert(cond bool) {
	if !cond {
		err := errors.New("assert fail")
		checkAndHandleError(err, err.Error(), verifyAction, _SKIP_LEVEL)
	}
}
