package request

import (
	"io"
	"strings"
	"fmt"
)

const bufferSize = 8

type Request struct {
	RequestLine RequestLine
	parserState ParserState
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type ParserState int
const (
	Initialized ParserState = iota
	Done
)

func (p ParserState) String() string {
	switch p {
	case Initialized:
		return "Initialized"
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
		parserState: 0,
	}

	for newRequest.parserState.String() != "Done" {
		if isFull(buf) {
			newBuf := make([]byte, len(buf) * 2, len(buf) * 2)
			copy(newBuf, buf)
			buf = newBuf
		}	

		bytesRead, err :=  reader.Read(buf[readToIndex:])

		if err == io.EOF {
			newRequest.parserState = 1
			break
		}

		readToIndex += bytesRead

		bytesParsed, err := newRequest.parse(buf[:readToIndex])

		if err != nil {
			return 	nil, fmt.Errorf("Error: %v", err)
		}
		slimBuf := make([]byte, len(buf), len(buf))
		copy(slimBuf, buf[:readToIndex])

		buf = slimBuf
		readToIndex -= bytesParsed	
	}

	return &newRequest, nil
}

func isFull(buf []byte) bool {
	for i := 0; i < len(buf); i++ {
		if buf[i] == 0  {
			return false
		}
	}

	return true
}

func parseRequestLine(line string) (*RequestLine, int, error) {
	lines := strings.Split(line, "\r\n")

	if (len(lines) < 2) {
		return nil, 0, nil
	}
	requestLineStr := lines[0]
	requestSplit := strings.Split(requestLineStr, " ")

	method := requestSplit[0]
	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return nil, 0, fmt.Errorf("Invalid method: %s", method)
		}
	}

	httpVersion := strings.Split(requestSplit[2], "/")
	if httpVersion[0] != "HTTP" {
		return nil, 0,  fmt.Errorf("Http Version is incorrect %s", httpVersion[0])
	}
	if httpVersion[1] != "1.1" {
		return nil, 0, fmt.Errorf("Http Version is incorrect %s", httpVersion[1])
	}

	requestLine := RequestLine {
		HttpVersion: httpVersion[1],
		RequestTarget: requestSplit[1],
		Method: requestSplit[0],
	}

	return &requestLine, len(line), nil
}

func (r *Request) parse(data []byte) (int, error) {
	if r.parserState.String() == "Initialized" {
		requestLine, bytesRead, err := parseRequestLine(string(data))
		if err != nil {
			return 0, fmt.Errorf("Error: %v", err)
		}

		if bytesRead == 0 {
			return 0, nil
		}

		// Update Request  Line field and change state to done
		r.parserState = 1
		r.RequestLine = *requestLine
		
		return len(data), nil

	} else if r.parserState.String() == "Done" {
		return 0, fmt.Errorf("Attempting read data in done state")
	} 

	return 0, fmt.Errorf("Unknown State\n")
}
