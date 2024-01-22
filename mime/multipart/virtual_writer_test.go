package multipart

import (
	"fishjam.com/go-library/flog"
	"fishjam.com/go-library/utils"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestUploadFileWithVirtualWriter(t *testing.T) {
	var (
		body []byte
	)
	uploadUrl := "http://127.0.0.1:9999/file/uploadWithMultipartReader"
	uploadFiles := []string{
		"virtual_writer.go",
		"virtual_writer_test.go",
		"F:\\ISO\\Windows\\Win10\\Win10_21H2_x64_CN_20220412.iso",
	}
	params := map[string]string{
		"key":          "value",
		"type":         "data",
		"keepFileName": "on",
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	client := http.Client{
		Transport: transport,
	}
	mpWrite := NewVirtualWriter()
	_ = mpWrite.SetBoundary("----WebKitFormBoundary2HwfBAoBw2hJ33gD")
	mpWrite.SetProgressCallback(func(part Part, err error, readCount, totalCount int64) {
		name := "<last>"
		if part != nil {
			name = part.Name()
		}

		flog.Debugf("on progress: part name=%s, err=%+v, readCount=%d, totalCount=%d, percent=%0.2f",
			name, err, readCount, totalCount, float64(readCount*100)/float64(totalCount))
	})

	defer mpWrite.Close()
	for key, val := range params {
		_ = utils.Verify(mpWrite.WriteField(key, val))
	}

	for _, uf := range uploadFiles {
		_ = utils.Verify(mpWrite.CreateFormFile("file", uf))
	}
	flog.Infof("content-length=%d", mpWrite.ContentLength())

	req := utils.VerifyWithResult(http.NewRequest(http.MethodPost, uploadUrl, mpWrite)).(*http.Request)

	req.Header.Set("Content-Type", mpWrite.FormDataContentType())
	resp, err := client.Do(req)
	if err == nil {
		body, err = io.ReadAll(resp.Body)
		defer resp.Body.Close()

		flog.Infof("response=%s, err=%+v", string(body), err)
	}

	//use sleep to wait, so can verify memory usage, or runtime.MemStats?
	if true {
		time.Sleep(time.Second * 30)
	}
}
