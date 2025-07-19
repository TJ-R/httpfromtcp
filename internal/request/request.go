package request

import (
	"io"
	"strings"
	"fmt"
	"errors"
	"bytes"
	"github.com/TJ-R/httpfromtcp/internal/headers"
	"strconv"
)

const bufferSize = 8

type Request struct {
	RequestLine RequestLine
	parserState ParserState
	Headers headers.Headers
	Body []byte
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type ParserState int
const (
	Initialized ParserState = iota
	RequestStateParsingHeaders
	RequestStateParsingBody
	Done
)

func (p ParserState) String() string {
	switch p {
	case Initialized:
		return "Initialized"
	case RequestStateParsingHeaders:
		return "Parsing Headers"
	case RequestStateParsingBody:
		return "Parsing Body"
	case Done:
		return "Done"
	default:
		return "Unknown"
	}
}


func RequestFromReader(reader io.Reader) (*Request, error) {
	buf := make([]byte, bufferSize, bufferSize)

	readToIndex := 0

	newRequest := Request {
		RequestLine: RequestLine{},		
		parserState: Initialized,
		Headers: headers.NewHeaders(),
	}

	for newRequest.parserState != Done {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf) * 2)
			copy(newBuf, buf)
			buf = newBuf
		}	

		bytesRead, err :=  reader.Read(buf[readToIndex:])
		if err != nil {
			if errors.Is(err, io.EOF) {
				if newRequest.parserState != Done {
					return nil, fmt.Errorf("Incomplete Request")
				}
				break
			}
			return nil, err
		}

		readToIndex += bytesRead
		bytesParsed, err := newRequest.parse(buf[:readToIndex])
		if err != nil {
			return 	nil, err
		}
		
		copy(buf, buf[bytesParsed:])


		readToIndex -= bytesParsed
	}

	return &newRequest, nil
}

func parseRequestLine(data []byte) (*RequestLine, int, error) {
	idx := bytes.Index(data, []byte("\r\n"))

	if idx == -1 {
		return nil, 0, nil
	}

	requestLineString := string(data[:idx])
	requestLine, err := requestLineFromString(requestLineString)
	if err != nil {
		return nil, 0, err
	}

	return requestLine, idx+2, nil

}

func requestLineFromString(line string) (*RequestLine, error) {
	requestSplit := strings.Split(line, " ")

	method := requestSplit[0]
	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return nil, fmt.Errorf("Invalid method: %s", method)
		}
	}

	httpVersion := strings.Split(requestSplit[2], "/")
	if httpVersion[0] != "HTTP" {
		return nil,  fmt.Errorf("Http Version is incorrect %s", httpVersion[0])
	}
	if httpVersion[1] != "1.1" {
		return nil, fmt.Errorf("Http Version is incorrect %s", httpVersion[1])
	}

	requestLine := RequestLine {
		HttpVersion: httpVersion[1],
		RequestTarget: requestSplit[1],
		Method: requestSplit[0],
	}

	return &requestLine,  nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.parserState {
	case Initialized:
		requestLine, bytesRead, err := parseRequestLine(data)
		if err != nil {
			return 0, fmt.Errorf("Error: %v", err)
		}

		if bytesRead == 0 {
			return 0, nil
		}

		// Update Request  Line field and change state to headers
		r.parserState = RequestStateParsingHeaders
		r.RequestLine = *requestLine
		return bytesRead, nil
	
	case RequestStateParsingHeaders:
		totalBytesParsed := 0

		for r.parserState != Done {
			n, err := r.parseSingle(data[totalBytesParsed:])	

			if err != nil {
				return 0, fmt.Errorf("Error: %v", err)
			}
			if n == 0 {
				return totalBytesParsed, nil
			}

			totalBytesParsed += n
		}

		return totalBytesParsed, nil

	case RequestStateParsingBody:
		contentLength := r.Get("Content-Length")
		contentLengthValue, err := strconv.Atoi(contentLength)
		if err != nil {
			return 0, fmt.Errorf("Error: Encountered an issue converting content-length to int")
		}

		if contentLength == "" || contentLengthValue == 0{
			r.parserState = Done
			return 0, nil
		} 

		r.Body = append(r.Body, data...)

		if len(data) == 0 && len(r.Body) < contentLengthValue {
			return 0, fmt.Errorf("Error: Body is less than content length")
		}
		
		if len(r.Body) > contentLengthValue {
			return 0, fmt.Errorf("Error: Body is greater than content length")
		}

		if len(r.Body) == contentLengthValue {
			r.parserState = Done
			return 0, nil
		}

		return len(data), nil

	case Done:
		return 0, fmt.Errorf("Attempting read data in done state")

	default:
		return 0, fmt.Errorf("Unknown State\n")
	}
}

func (r *Request) parseSingle(data []byte) (int, error) {
	bytesParsed, done, err := r.Headers.Parse(data)

	if err != nil {
		return 0, fmt.Errorf("Error: %v", err)
	}

	if done {

		if r.Get("Content-Length") == "" || r.Get("Content-Length") == "0" {
			r.parserState = Done
			return bytesParsed, nil
		} else {
			r.parserState = RequestStateParsingBody
			return bytesParsed+2, nil
		}
	}

	if bytesParsed == 0 {
		return 0, nil
	}

	return bytesParsed, nil
}

func (r *Request) Get(key string) string {
	value, ok := r.Headers[strings.ToLower(key)]
	if !ok {
		return ""
	}

	return value
}
