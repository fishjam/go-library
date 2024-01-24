package utils

import (
	"github.com/fishjam/go-library/flog"
	"os"
	"testing"
)

func TestVerify(t *testing.T) {
	//this is a simple test functions for demonstrate how to use VerifyXxx functions
	file := VerifyWithResult[*os.File](os.Open("not_exist_file"))
	flog.Warnf("file=%+v", file)

	if file != nil {
		defer func() {
			_ = Verify(file.Close())
		}()
		//handle file
	}
}
