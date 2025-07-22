package response

import (
	"io"
	"fmt"
	"github.com/TJ-R/httpfromtcp/internal/headers"
)

type StatusCode int 

const (
	StatusOk StatusCode = iota
	StatusClientError
	StatusServerError
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	message := ""

	switch statusCode {
	case StatusOk:
		message = "HTTP/1.1 200 OK\r\n"	
	case StatusClientError: 
		message = "HTTP/1.1 400 Bad Request\r\n"	
	case StatusServerError:
		message = "HTTP/1.1 500 Internal Server Error\r\n"	
	default:
		message = "HTTP/1.1\r\n"	
	}

	_, err := w.Write([]byte(message))
	if err != nil {
		return err
	}

	return nil
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	headers["content-length"] = fmt.Sprintf("%v", contentLen)
	headers["connection"] = "close"
	headers["content-type"] = "text/plain"

	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {
		_, err := w.Write([]byte(k + ": " + v + "\r\n"))
		if err != nil {
			return err
		}
	}
	
	w.Write([]byte("\r\n\r\n"))
	return nil
}
