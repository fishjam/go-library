package ioext

import (
	"bytes"
	"fmt"
	"github.com/fishjam/go-library/debugutil"
	"github.com/fishjam/go-library/flog"
	"io"
	"net/http"
	"reflect"
)

// https://blog.csdn.net/zhang197093/article/details/127407838

// 可重复读取的 Reader, 可用于 http 中再次发送请求等.
type RepeatableReader struct {
	orgReader  io.Reader
	newReader  io.Reader
	readerFunc ReaderFunc
	orgLength  int64
}

func NewRepeatableReader(reader io.Reader) *RepeatableReader {
	flog.Debugf("reader type=%s", reflect.TypeOf(reader).String())
	readerFunc, length, err := getBodyReaderAndContentLength(reader)
	if err != nil {
		return nil
	}

	newReader := debugutil.VerifyWithResult(readerFunc())
	return &RepeatableReader{
		orgReader:  reader,
		readerFunc: readerFunc,
		newReader:  newReader,
		orgLength:  length,
	}
}

func (rr *RepeatableReader) Read(p []byte) (n int, err error) {
	return rr.newReader.Read(p)
}

func (rr *RepeatableReader) Close() error {
	if closer, ok := rr.orgReader.(io.Closer); ok {
		flog.Debugf("enter RepeatableReader.Close")
		return closer.Close()
	}
	return nil
}

func (rr *RepeatableReader) Reset() error {
	newReader, err := rr.readerFunc()
	if err != nil {
		return err
	}
	rr.newReader = newReader
	return err
}

type ReaderFunc func() (io.Reader, error)

// LenReader is an interface implemented by many in-memory io.Reader's. Used
// for automatically sending the right Content-Length header when possible.
type LenReader interface {
	Len() int
}

func getBodyReaderAndContentLength(rawBody interface{}) (ReaderFunc, int64, error) {
	var bodyReader ReaderFunc
	var contentLength int64

	flog.Debugf("readerFunc rawBody type=%s", reflect.TypeOf(rawBody).String())

	switch body := rawBody.(type) {
	// If they gave us a function already, great! Use it.
	case ReaderFunc:
		bodyReader = body
		tmp, err := body()
		if err != nil {
			return nil, 0, err
		}
		if lr, ok := tmp.(LenReader); ok {
			contentLength = int64(lr.Len())
		}
		if c, ok := tmp.(io.Closer); ok {
			c.Close()
		}

	case func() (io.Reader, error):
		bodyReader = body
		tmp, err := body()
		if err != nil {
			return nil, 0, err
		}
		if lr, ok := tmp.(LenReader); ok {
			contentLength = int64(lr.Len())
		}
		if c, ok := tmp.(io.Closer); ok {
			c.Close()
		}

	// If a regular byte slice, we can read it over and over via new
	// readers
	case []byte:
		buf := body
		bodyReader = func() (io.Reader, error) {
			return bytes.NewReader(buf), nil
		}
		contentLength = int64(len(buf))

	// If a bytes.Buffer we can read the underlying byte slice over and
	// over
	case *bytes.Buffer:
		buf := body
		bodyReader = func() (io.Reader, error) {
			return bytes.NewReader(buf.Bytes()), nil
		}
		contentLength = int64(buf.Len())

	// We prioritize *bytes.Reader here because we don't really want to
	// deal with it seeking so want it to match here instead of the
	// io.ReadSeeker case.
	case *bytes.Reader:
		buf, err := io.ReadAll(body)
		if err != nil {
			return nil, 0, err
		}
		bodyReader = func() (io.Reader, error) {
			return bytes.NewReader(buf), nil
		}
		contentLength = int64(len(buf))

	// Compat case
	case io.ReadSeeker:
		raw := body
		bodyReader = func() (io.Reader, error) {
			_, err := raw.Seek(0, 0)
			return io.NopCloser(raw), err
		}
		if lr, ok := raw.(LenReader); ok {
			contentLength = int64(lr.Len())
		}

	// Read all in so we can reset
	case io.Reader:
		buf, err := io.ReadAll(body)
		if err != nil {
			return nil, 0, err
		}
		if len(buf) == 0 {
			bodyReader = func() (io.Reader, error) {
				return http.NoBody, nil
			}
			contentLength = 0
		} else {
			bodyReader = func() (io.Reader, error) {
				return bytes.NewReader(buf), nil
			}
			contentLength = int64(len(buf))
		}

	// No body provided, nothing to do
	case nil:

	// Unrecognized type
	default:
		return nil, 0, fmt.Errorf("cannot handle type %T", rawBody)
	}
	return bodyReader, contentLength, nil
}
