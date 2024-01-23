package multipart

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
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

// please Notice: after this test case run, will create upload folder, and upload some files into it,
// can compare the content
func TestUploadFileWithVirtualWriter(t *testing.T) {

	uploadTempFolder, err := os.MkdirTemp(os.TempDir(), "virtual_writer")
	t.Logf("uploadTempFolder=%s", uploadTempFolder)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mpReader, err := r.MultipartReader()
		//t.Logf("mpReader type is %+v, err=%+v", reflect.TypeOf(mpReader), err)

		uploadFields := make(map[string]string)
		err = os.MkdirAll(uploadTempFolder, 0755)
		assert.Nil(t, err, fmt.Sprintf("create upload temp folder:%s", uploadTempFolder))

		for {
			part, err := mpReader.NextPart()
			//必须判断
			if err == io.EOF {
				//t.Logf("nextPart end,")
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
			outFile, err := os.Create(filePath)
			assert.Nil(t, err, fmt.Sprintf("create upload file: %s", filePath))
			defer outFile.Close()

			_, err = io.Copy(outFile, part)
		}
		marshalResult, err := json.Marshal(uploadFields)
		w.WriteHeader(http.StatusOK)
		w.Write(marshalResult)
	}))

	defer ts.Close()
	t.Logf("ts.URL=%s", ts.URL)

	uploadResultExpected := make(map[string]string)

	uploadUrl := fmt.Sprintf("%s%s", ts.URL, "/upload")
	uploadFiles := []string{
		"virtual_writer.go",
		"virtual_writer_test.go",
		//"not_exist",
		//"F:\\ISO\\Windows\\Win10\\Win10_21H2_x64_CN_20220412.iso",
		//"D:\\Git\\git-bash.exe",
	}
	params := map[string]string{
		"key":          "value",
		"type":         "data",
		"keepFileName": "on",
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	localProxyPort := 0
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
	//_ = mpWrite.SetBoundary("----WebKitFormBoundary2HwfBAoBw2hJ33gD")
	mpWrite.SetProgressCallback(func(part VirtualPart, err error, readCount, totalCount int64) {
		name := "<last>"
		if part != nil {
			name = part.Name()
		}
		_ = name

		//t.Logf("on progress: part name=%s, err=%+v, readCount=%d, totalCount=%d, percent=%0.2f",
		//	name, err, readCount, totalCount, float64(readCount*100)/float64(totalCount))
		assert.True(t, readCount <= totalCount, "progress")
	})

	defer mpWrite.Close()
	for key, val := range params {
		_ = mpWrite.WriteField(key, val)
		uploadResultExpected[key] = val
	}

	fileIndex := 0
	for _, uf := range uploadFiles {
		fieldName := fmt.Sprintf("file%d", fileIndex)
		err = mpWrite.CreateFormFile(fieldName, uf)
		assert.Nil(t, err, "CreateFormFile")
		uploadResultExpected[fieldName] = filepath.Base(uf)
		fileIndex++
	}
	t.Logf("content-length=%d", mpWrite.ContentLength())

	req, err := http.NewRequest(http.MethodPost, uploadUrl, mpWrite)

	req.Header.Set("Content-Type", mpWrite.FormDataContentType())
	resp, err := client.Do(req)
	assert.Nil(t, err, "client.Do should successful")

	if err == nil {
		body, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()

		var uploadResponse map[string]string
		err = json.Unmarshal(body, &uploadResponse)
		assert.Nil(t, err, "unmarshal result successful")
		t.Logf("response=%s, err=%+v", string(body), err)

		assert.Equal(t, uploadResultExpected, uploadResponse, "upload result")
	}

	for idx, uf := range uploadFiles {
		srcSum := fileCheckSum(uf)

		dstPath := path.Join(uploadTempFolder, filepath.Base(uf))
		dstSum := fileCheckSum(dstPath)
		assert.Equal(t, srcSum, dstSum, fmt.Sprintf("compare[%d] %s <=> %s", idx, uf, dstPath))
	}

	//use sleep to wait, so can verify memory usage, or runtime.MemStats?
	if false {
		time.Sleep(time.Second * 30)
	}
	os.RemoveAll(uploadTempFolder)
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

	// 格式化为16进制字符串
	return fmt.Sprintf("%x", h.Sum(nil))
}
