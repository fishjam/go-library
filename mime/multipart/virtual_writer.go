package multipart

import (
	"bufio"
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// OnProgressCallback provide progress callback when do real POST, it's useful when uploading large files.
//
// Notice: the part is nil and err is EOF when send last end boundary.
type OnProgressCallback func(part VirtualPart, err error, readCount, totalCount int64)

// VirtualPart handle all the part's content and provide detail information in OnProgressCallback
type VirtualPart interface {
	// Name returns current field name
	Name() string

	// Len returns current part length, include boundary
	Len() int64

	// Remain returns current remain length while reading
	Remain() int64

	// Close closes the part(FilePart)
	Close() error

	read(p []byte) (n int, err error)
}

type fieldPart struct {
	fieldName   string
	fieldValue  string
	fieldLength int64
	readOffset  int64
}

func (fp *fieldPart) Name() string {
	return fp.fieldName
}

func (fp *fieldPart) Len() int64 {
	return fp.fieldLength
}

func (fp *fieldPart) read(p []byte) (n int, err error) {
	reader := bytes.NewReader([]byte(fp.fieldValue[fp.readOffset:]))
	bufReader := bufio.NewReader(reader)
	n, err = bufReader.Read(p)
	if err == io.EOF {
		err = nil
	}
	fp.readOffset += int64(n)
	return
}
func (fp *fieldPart) Remain() int64 {
	return fp.fieldLength - fp.readOffset
}

func (fp *fieldPart) Close() error {
	return nil
}

type filePart struct {
	fieldValue  string
	fieldLength int64
	readOffset  int64
	filePath    string
	fileSize    int64
	file        *os.File
	once        sync.Once
}

func (fp *filePart) Name() string {
	return fp.filePath
}

func (fp *filePart) Len() int64 {
	//the last 2 is last \r\n after file content
	return fp.fieldLength + fp.fileSize + 2
}

func (fp *filePart) read(p []byte) (n int, err error) {
	var (
		nField    int
		nFile     int
		nLastCrLf int
	)
	fp.once.Do(func() {
		//open file
		fp.file, err = os.Open(fp.filePath)
		if err != nil {
			//open file fail, example: delete file after CreateFormFile
			return
		}
	})
	if err != nil {
		//once.Do error
		return 0, err
	}
	if fp.readOffset < fp.fieldLength {
		// read from field
		reader := bytes.NewReader([]byte(fp.fieldValue[fp.readOffset:]))
		nField, err = reader.Read(p)
		fp.readOffset += int64(nField)
		if err == io.EOF {
			err = nil
		}
	}
	if fp.readOffset >= fp.fieldLength {
		//read from file
		fileOffset := fp.readOffset - fp.fieldLength
		nFile, err = fp.file.ReadAt(p[nField:], fileOffset)
		if err == io.EOF {
			err = nil
			//after read file end, then append \r\n
			reader := bytes.NewReader([]byte("\r\n"))
			nLastCrLf, err = reader.Read(p[(nField + nFile):])
			fp.readOffset += int64(nLastCrLf) // 2 char
		}
		if err == nil {
			//total read count
			fp.readOffset += int64(nFile)
		}
	}
	//current read count
	n = nField + nFile + nLastCrLf
	return n, err
}

func (fp *filePart) Remain() int64 {
	return fp.Len() - fp.readOffset
}

func (fp *filePart) Close() (err error) {
	if fp.file != nil {
		err = fp.file.Close()
		fp.file = nil
	}
	return err
}

// VirtualWriter is similar as builtin [mime/multipart.writer],
// but can support upload lots of files (4G+) with little memory consume.
type VirtualWriter struct {
	closeAfterRead bool
	boundary       string
	parts          []VirtualPart
	finished       bool
	readPartIndex  int
	readCount      int64
	totalCount     int64
	callback       OnProgressCallback
}

// NewVirtualWriter returns a new multipart writer with a random boundary,
func NewVirtualWriter() *VirtualWriter {
	boundary := randomBoundary()
	return &VirtualWriter{
		closeAfterRead: true,
		boundary:       boundary,
		parts:          make([]VirtualPart, 0),
		readPartIndex:  0,
		readCount:      0,
		totalCount:     int64(len(boundary) + 6), //init total count is for last boundary(--%s--\r\n)
	}
}

// SetCloseAfterRead close the part immediately after read the part, it's used in file part, default is true
func (vw *VirtualWriter) SetCloseAfterRead(closeAfterRead bool) {
	vw.closeAfterRead = closeAfterRead
}

// SetBoundary overrides the VirtualWriter's default randomly-generated
// boundary separator with an explicit value.
//
// Copy from built-in multipart.Writer, and modify for totalCount
func (vw *VirtualWriter) SetBoundary(boundary string) error {
	if len(vw.parts) > 0 {
		return errors.New("mime: SetBoundary called after write")
	}

	// rfc2046#section-5.1.1
	if len(boundary) < 1 || len(boundary) > 70 {
		return errors.New("mime: invalid boundary length")
	}

	end := len(boundary) - 1
	for i, b := range boundary {
		if 'A' <= b && b <= 'Z' || 'a' <= b && b <= 'z' || '0' <= b && b <= '9' {
			continue
		}
		switch b {
		case '\'', '(', ')', '+', '_', ',', '-', '.', '/', ':', '=', '?':
			continue
		case ' ':
			if i != end {
				continue
			}
		}
		return errors.New("mime: invalid boundary character")
	}
	vw.boundary = boundary

	vw.totalCount = int64(len(boundary) + 6) //init total count is for last boundary(--%s--\r\n)
	return nil
}

// FormDataContentType returns the Content-Type for an HTTP
// multipart/form-data with this Writer's Boundary.
//
// Copy from built-in multipart.Writer #FormDataContentType
func (vw *VirtualWriter) FormDataContentType() string {
	b := vw.boundary
	// We must quote the boundary if it contains any of the
	// tspecials characters defined by RFC 2045, or space.
	if strings.ContainsAny(b, `()<>@,;:\"/[]?= `) {
		b = `"` + b + `"`
	}
	return "multipart/form-data; boundary=" + b
}

func randomBoundary() string {
	var buf [30]byte
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", buf[:])
}

var quoteEscaper = strings.NewReplacer("\\", "\\\\", `"`, "\\\"")

func escapeQuotes(s string) string {
	return quoteEscaper.Replace(s)
}

// SetProgressCallback set the callback function
func (vw *VirtualWriter) SetProgressCallback(callback OnProgressCallback) {
	vw.callback = callback
}

// CreateFormFile creates a new form-data header with the provided field name and file name.
// But it just remains the file information, will not read the file contents to memory until actual POST occurs.
func (vw *VirtualWriter) CreateFormFile(fieldName, filePath string) error {
	stat, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	fieldValue := fmt.Sprintf("--%s\r\nContent-Disposition: form-data; name=\"%s\"; filename=\"%s\"\r\n"+
		"Content-Type: application/octet-stream\r\n\r\n",
		vw.boundary, escapeQuotes(fieldName), filepath.Base(filePath))

	part := filePart{
		fieldValue:  fieldValue,
		filePath:    filePath,
		fieldLength: int64(len(fieldValue)),
		fileSize:    stat.Size(),
		readOffset:  0,
	}
	vw.parts = append(vw.parts, &part)
	vw.totalCount += part.Len()
	return nil
}

// WriteField creates a new form-data header with the provided field name and value
func (vw *VirtualWriter) WriteField(fieldName, value string) error {
	fieldVal := fmt.Sprintf("--%s\r\nContent-Disposition: form-data; name=\"%s\"\r\n\r\n%s\r\n",
		vw.boundary,
		escapeQuotes(fieldName), value)
	part := fieldPart{
		fieldName:   fieldName,
		fieldValue:  fieldVal,
		fieldLength: int64(len(fieldVal)),
		readOffset:  0,
	}
	vw.parts = append(vw.parts, &part)
	vw.totalCount += part.Len()
	return nil
}

// Read is used for Reader function(example: body parameter for http.NewRequest),
// user should NOT call it directly.
func (vw *VirtualWriter) Read(p []byte) (nRead int, err error) {
	var (
		nReadLastBoundary int
	)

	if vw.readPartIndex < len(vw.parts) {
		part := vw.parts[vw.readPartIndex]
		nRead, err = part.read(p) //p[nb:])
		//log.Printf("idx[%d], part read, nRead=%d, remain=%d, err=%+v",
		//	vw.readPartIndex, nRead, part.Remain(), err)

		if err == nil {
			//TODO: read than available data, file change after add ?
			//return -1, errors.New("read more data than available")
			if part.Remain() == 0 {
				if vw.closeAfterRead {
					_ = part.Close()
				}
				//read all current part data, will try read next
				vw.readPartIndex++
			}
			vw.readCount += int64(nRead)
		}
		if vw.callback != nil {
			vw.callback(part, err, vw.readCount, vw.totalCount)
		}
	}
	if vw.readPartIndex >= len(vw.parts) {
		//already read all part's data
		strLastBoundary := fmt.Sprintf("--%s--\r\n", vw.boundary)
		reader := bytes.NewReader([]byte(strLastBoundary))
		nReadLastBoundary, err = reader.Read(p[nRead:])
		if err == nil {
			nRead += nReadLastBoundary
			err = io.EOF
		}
		vw.readCount += int64(nReadLastBoundary)
		if vw.callback != nil {
			vw.callback(nil, err, vw.readCount, vw.totalCount)
		}
	}
	return nRead, err
}

// ContentLength return content length for all parts, it should same as "Content-Length" in http header
func (vw *VirtualWriter) ContentLength() int64 {
	return vw.totalCount
}

// Close closes all the parts
func (vw *VirtualWriter) Close() error {
	var err error
	for _, part := range vw.parts {
		pErr := part.Close()
		if err == nil && pErr != nil {
			//just return first error
			err = pErr
		}
	}
	vw.readCount = 0
	vw.totalCount = 0
	return err
}
