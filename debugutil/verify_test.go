package debugutil

import (
	"os"
	"testing"
)

// TestVerify this is a simple test functions for demonstrate how to use VerifyXxx functions.
//
// example: open a file should exist(local config file),
// if it not exists, then it's code error or CI/CD error, not runtime error.
func TestVerify(t *testing.T) {
	file := VerifyWithResult[*os.File](os.Open("should_exist_conf_file"))

	defer func() {
		//Notice: when try to close a nil(*os.File), error with "invalid argument"
		_ = Verify(file.Close())
	}()
}

func someFunReturnValue() bool {
	return false
}

func TestAssert(t *testing.T) {
	Assert(someFunReturnValue())
}
