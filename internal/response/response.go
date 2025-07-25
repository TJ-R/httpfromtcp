package response

import (
	"fmt"
	"io"
	"maps"
	"github.com/TJ-R/httpfromtcp/internal/headers"
)

type StatusCode int 
type WriterState int

const (
	StatusOk StatusCode = 200
	StatusClientError   = 400
	StatusServerError   = 500
)

const (
	WritingStatus WriterState = iota
	WritingHeaders
	WritingBody
)


type Writer struct {
	W io. Writer
	writerState WriterState
	StatusCode StatusCode
	Headers    headers.Headers
	Body       []byte
}

func (writer *Writer) GetStatusLine() {

}

func (writer *Writer) WriteStatusLine(statusCode StatusCode) error {
	writer.StatusCode = statusCode

	statusReason := ""
	switch statusCode {
	case StatusOk:
		statusReason = "OK"
	case StatusClientError:
		statusReason = "Bad Request"
	case StatusServerError:
		statusReason = "Internal Server Error"
	default:
		statusReason = "Unknown Status Code"
	}

	_, err := writer.W.Write([]byte(fmt.Sprintf("HTTP/1.1 %d %s\r\n", writer.StatusCode, statusReason)))
	if err != nil {
		return err
	}

	writer.writerState = WritingHeaders

	return nil
}

func (writer *Writer) WriteHeaders(newHeaders headers.Headers) error {
	if writer.writerState != WritingHeaders {
		return fmt.Errorf("Writing Headers before StatusLine")
	}

	writer.Headers = headers.NewHeaders() 

	maps.Copy(writer.Headers, newHeaders)
	
	for k, v := range writer.Headers {
		_, err := writer.W.Write([]byte(k + ": " + v + "\r\n"))
		if err != nil {
			return err
		}
	}
	
	_, err := writer.W.Write([]byte("\r\n")) 
	if err != nil {
		return err
	}

	writer.writerState = WritingBody
	return nil
}

func (writer *Writer) WriteBody(p []byte) error {
	if writer.writerState != WritingBody {
		return fmt.Errorf("Writing Body before Headers")
	}

	_, err := writer.W.Write(p)
	if err != nil {
		return err
	}

	return nil
}

func (writer *Writer) WriteChunkedBody(p []byte) (int, error) {
	if writer.writerState != WritingBody {
		return 0, fmt.Errorf("Writing Body before Headers")
	}

	totalBytes := 0
	n, err := writer.W.Write([]byte(fmt.Sprintf("%x\r\n", len(p))))
	totalBytes += n
	if err != nil {
		return 0, err
	}

	n, err = writer.W.Write(p)
	totalBytes += n
	if err != nil {
		return 0, err
	}

	n, err = writer.W.Write([]byte("\r\n"))
	totalBytes += n
	if err != nil {
		return 0, err
	}

	return n, nil
}

func (writer *Writer) WriteChunkedBodyDone() (int, error) {
	_, err := writer.W.Write([]byte("0\r\n\r\n"))
	if err != nil {
		return 0, err
	}

	return 0, nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	headers["content-length"] = fmt.Sprintf("%v", contentLen)
	headers["connection"] = "close"
	headers["content-type"] = "text/plain"

	return headers
}


