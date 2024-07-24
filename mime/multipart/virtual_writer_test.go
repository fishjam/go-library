package multipart

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fishjam/go-library/debugutil"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"
)

var uploadFiles = []string{
	"virtual_writer.go",
	"virtual_writer_test.go",
	//"not_exist",
	// some large file
	//"F:\\ISO\\Windows\\Win10\\Win10_21H2_x64_CN_20220412.iso",
}

// TestUploadFileWithVirtualWriter
// after this test case run, will create upload folder, and upload some files into it,
// then compare the source and target file's md5
func TestUploadFilesWithVirtualWriter(t *testing.T) {
	//local fiddler proxy port, if not 0(example: 8888), then can use local fiddle to monitor network data
	localProxyPort := 0

	//this UT will remove all upload files after test, if you want remains the uploaded files to compare,
	//can set removeUploadFiles to false
	removeUploadFiles := true

	uploadTempFolder := debugutil.VerifyWithResult(os.MkdirTemp(os.TempDir(), "virtual_writer"))
	t.Logf("uploadTempFolder=%s", uploadTempFolder)
	if removeUploadFiles {
		defer func() {
			_ = debugutil.Verify(os.RemoveAll(uploadTempFolder))
		}()
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mpReader := debugutil.VerifyWithResult(r.MultipartReader())

		uploadFields := make(map[string]string)
		_ = debugutil.Verify(os.MkdirAll(uploadTempFolder, 0755))

		for {
			part, err := mpReader.NextPart()
			//part, err := utils.VerifyWithResultEx[*multipart.Part](mpReader.NextPart())
			if err == io.EOF {
				break
			}

			name := part.FileName()
			if name == "" {
				buf := make([]byte, 1024)
				n, err := part.Read(buf)
				//t.Logf("part.Read error:%+v, n=%d", err, n)
				if err != nil { // 等于 EOF 时表示读取完毕,之后 buf 中的才是结果,
					uploadFields[part.FormName()] = string(buf[0:n])
				} else {
					uploadFields[part.FormName()] = string(buf[0:n]) //TODO: alloc buf?
				}
				continue
			}
			uploadFields[part.FormName()] = part.FileName()

			filename := part.FileName()
			filePath := path.Join(uploadTempFolder, filename)
			outFile := debugutil.VerifyWithResult(os.Create(filePath))
			defer outFile.Close()

			_ = debugutil.VerifyWithResult(io.Copy(outFile, part))
		}
		marshalResult := debugutil.VerifyWithResult(json.Marshal(uploadFields))
		w.WriteHeader(http.StatusOK)
		_ = debugutil.VerifyWithResult(w.Write(marshalResult))
	}))

	defer ts.Close()
	t.Logf("ts.URL=%s", ts.URL)

	uploadResultExpected := make(map[string]string)

	uploadUrl := fmt.Sprintf("%s%s", ts.URL, "/upload")

	params := map[string]string{
		"key":          "value",
		"type":         "data",
		"keepFileName": "on",
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	if localProxyPort != 0 {
		proxyAddr := fmt.Sprintf("http://127.0.0.1:%d", localProxyPort)
		proxyUrl, _ := url.Parse(proxyAddr)
		proxyFunc := http.ProxyURL(proxyUrl)
		transport.Proxy = proxyFunc
	}

	client := http.Client{
		Transport: transport,
	}
	mpWrite := NewVirtualWriter()
	defer func() {
		_ = debugutil.Verify(mpWrite.Close())
	}()

	//_ = mpWrite.SetBoundary("----WebKitFormBoundary2HwfBAoBw2hJ33gD")
	mpWrite.SetProgressCallback(func(part VirtualPart, err error, readCount, totalCount int64) {
		name := "<last>"
		if part != nil {
			name = part.Name()
		}
		_ = name

		t.Logf("on progress: part name=%s, err=%+v, readCount=%d, totalCount=%d, percent=%0.2f",
			name, err, readCount, totalCount, float64(readCount*100)/float64(totalCount))
		debugutil.GoAssertTrue(t, readCount <= totalCount, "progress")
	})

	for key, val := range params {
		_ = mpWrite.WriteField(key, val)
		uploadResultExpected[key] = val
	}

	fileIndex := 0
	for _, uf := range uploadFiles {
		fieldName := fmt.Sprintf("file%d", fileIndex)
		if err := debugutil.Verify(mpWrite.CreateFormFile(fieldName, uf)); err != nil {
			t.Logf("CreateFormFile error, %+v", err)
			//in real code, should handle this error and maybe return

			//return
		}
		uploadResultExpected[fieldName] = filepath.Base(uf)
		fileIndex++
	}
	t.Logf("content-length=%d", mpWrite.ContentLength())

	req, err := http.NewRequest(http.MethodPost, uploadUrl, mpWrite)

	req.Header.Set("Content-Type", mpWrite.FormDataContentType())
	resp, err := client.Do(req)
	debugutil.GoAssertTrue(t, err == nil, "client.Do should successful")

	if err == nil {
		body := debugutil.VerifyWithResult[[]byte](io.ReadAll(resp.Body))
		defer resp.Body.Close()

		var uploadResponse map[string]string
		_ = debugutil.Verify(json.Unmarshal(body, &uploadResponse))
		t.Logf("response=%s, err=%+v", string(body), err)

		debugutil.GoAssertEqual(t, uploadResultExpected, uploadResponse, "upload result")
	}

	//check the uploaded files, should same as original files by check sum(MD5)
	for idx, uf := range uploadFiles {
		srcSum := fileCheckSum(uf)

		dstPath := path.Join(uploadTempFolder, filepath.Base(uf))
		dstSum := fileCheckSum(dstPath)

		debugutil.GoAssertEqual(t, srcSum, dstSum, fmt.Sprintf("compare[%d] %s <=> %s", idx, uf, dstPath))
	}

	//use sleep to wait, so can verify memory usage, or runtime.MemStats?
	if false {
		time.Sleep(time.Second * 30)
	}
}

func TestVirtualWriterSeek(t *testing.T) {
	mpWrite := NewVirtualWriter()
	mpWrite.SetCloseAfterRead(false)

	pos, err := mpWrite.Seek(9999, io.SeekStart)
	debugutil.GoAssertTrue(t, pos == 0 && errors.Is(err, ErrWrongParam), "seek start wrong")

	fileIndex := 0
	for _, uf := range uploadFiles {
		fieldName := fmt.Sprintf("file%d", fileIndex)
		if err := debugutil.Verify(mpWrite.CreateFormFile(fieldName, uf)); err != nil {
			t.Logf("CreateFormFile error, %+v", err)
		}
		fileIndex++
	}

	Cases := []struct {
		offset int64
		whence int
	}{
		{0, io.SeekStart},
		//{0, io.SeekCurrent},
		//{0, io.SeekEnd},
		//
		//{mpWrite.totalCount - 1, io.SeekStart},
	}

	writer := bytes.NewBuffer(nil)
	lenght, err := io.Copy(writer, mpWrite)
	t.Logf("after io.Copy, lenght=%d, err=%+v", lenght, err)

	for _, testCase := range Cases {
		t.Logf("before mpWrite readCount=%d, totalCount=%d, readPartIndex=%d, len(parts)=%d",
			mpWrite.readCount, mpWrite.totalCount, mpWrite.readPartIndex, len(mpWrite.parts))

		debugutil.VerifyWithResult(mpWrite.Seek(testCase.offset, testCase.whence))

		t.Logf("after mpWrite readCount=%d, totalCount=%d, readPartIndex=%d, len(parts)=%d",
			mpWrite.readCount, mpWrite.totalCount, mpWrite.readPartIndex, len(mpWrite.parts))
	}
	lenght, err = io.Copy(writer, mpWrite)
	t.Logf("after second io.Copy, lenght=%d, err=%+v", lenght, err)
}

func fileCheckSum(fileName string) string {
	f, err := os.Open(fileName)
	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	h := md5.New()
	//h := sha256.New()
	//h := sha1.New()
	//h := sha512.New()

	if _, err := io.Copy(h, f); err != nil {
		log.Fatal(err)
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}
