package utils

import (
	"fmt"
	"github.com/fishjam/go-library/flog"
	"reflect"
	"testing"
)

func errorWithInfo(t *testing.T, msg string, skip int) {

	if true {
		fileName, lineNo, funName := flog.GetCallStackInfo(skip)
		flog.Warnf("%s:%d: (%s) FAIL, msg=%s, ", fileName, lineNo, funName, msg)
		//panic(msg)
		t.Errorf("") //"FAIL: %s:%d, msg=%s", fname, lineno, msg)
	} else {
		t.Helper()
		t.Errorf("FAIL: msg=%s", msg)
	}

}

func GoAssertTrue(t *testing.T, bAssert bool, msg string) {
	//t.Helper()
	if !bAssert {
		errorWithInfo(t, msg, 3)
	}
}

// TODO: 和 assert.equal 比较?
func GoAssertEqual(t *testing.T, expected, actual any, msg string) {
	t.Helper()
	isEqual := false
	if expected == nil || actual == nil {
		isEqual = expected == actual // 由于这里必然有一个是 nil, 因此可以直接 ==
	} else if isComparable := reflect.TypeOf(expected).Comparable() && reflect.TypeOf(actual).Comparable(); isComparable {
		//判断是否支持比较, 否则会 panic: runtime error: comparing uncomparable type xxx
		//https://www.jb51.net/article/271916.htm
		isEqual = expected == actual
	} else {
		isEqual = reflect.DeepEqual(expected, actual)
	}

	if !isEqual {
		msg := fmt.Sprintf("%s: %s(\"%+v\") != %s(\"%+v\")", msg,
			reflect.TypeOf(expected).String(), expected,
			reflect.TypeOf(actual).String(), actual)
		errorWithInfo(t, msg, 3)
	}
}
