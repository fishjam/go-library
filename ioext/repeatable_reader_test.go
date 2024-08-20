package ioext

import (
	"bytes"
	"github.com/fishjam/go-library/debugutil"
	"github.com/fishjam/go-library/flog"
	multipart "github.com/fishjam/go-library/mime/multipart"
	"io"
	"testing"
)

func TestRepeatableReader(t *testing.T) {
	type CheckFunc func(rawData []byte, length int64)

	Cases := []struct {
		reader    io.Reader
		checkFunc CheckFunc
	}{
		//{
		//	strings.NewReader("fishjam"),
		//	func(rawData []byte, length int64) {
		//		debugutil.GoAssertEqual(t, "fishjam", string(rawData), "strings.NewReader")
		//	},
		//},
		//{
		//	debugutil.VerifyWithResult(os.Open("repeatable_reader_test.go")),
		//	func(rawData []byte, length int64) {
		//		fileinfo := debugutil.VerifyWithResult(os.Stat("repeatable_reader_test.go"))
		//		debugutil.GoAssertEqual(t, length, fileinfo.Size(), "read file size")
		//	},
		//},
		{
			func() io.Reader {
				mpWrite := multipart.NewVirtualWriter()
				mpWrite.SetCloseAfterRead(false)
				mpWrite.CreateFormFile("file1", "repeatable_reader.go")
				mpWrite.CreateFormFile("file2", "repeatable_reader_test.go")
				return mpWrite
			}(),
			func(rawData []byte, length int64) {
				debugutil.GoAssertTrue(t, length > 0, "VirtualWriter length")
			},
		},
	}

	for _, testCase := range Cases {
		repeatableReader := NewRepeatableReader(testCase.reader)
		for i := 0; i < 2; i++ {
			flog.Infof("read index %d", i)
			buffer := bytes.NewBuffer(nil)
			written := debugutil.VerifyWithResult(io.Copy(buffer, repeatableReader))
			testCase.checkFunc(buffer.Bytes(), written)

			subStrLen := written
			if subStrLen > 200 {
				subStrLen = 200
			}
			flog.Infof("read all , length=%d, contents=%s", written, buffer.String()[0:subStrLen])
			_ = debugutil.Verify(repeatableReader.Reset())
		}

		_ = debugutil.Verify(repeatableReader.Close())
	}
}
