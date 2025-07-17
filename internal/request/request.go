package request

import (
	"io"
	"strings"
	"fmt"
	"errors"
	"bytes"
	"github.com/TJ-R/httpfromtcp/internal/headers"
)

const bufferSize = 8

type Request struct {
	RequestLine RequestLine
	parserState ParserState
	Headers headers.Headers
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
	Done
)

func (p ParserState) String() string {
	switch p {
	case Initialized:
		return "Initialized"
	case RequestStateParsingHeaders:
		return "Parsing Headers"
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
		fmt.Printf("Read To Index: %v\n", readToIndex)
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

	return requestLine, idx + 2, nil

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
		return len(data), nil
	
	case RequestStateParsingHeaders:
		fmt.Printf("Len of data: %v\n", len(data))
		totalBytesParsed := 0

		for r.parserState != Done {
			n, err := r.parseSingle(data[totalBytesParsed:])	

			//fmt.Printf("Number of bytes parsed: %v\n", n)
			if err != nil {
				return 0, fmt.Errorf("Error: %v", err)
			}
			if n == 0 {
				return totalBytesParsed, nil
			}

			totalBytesParsed += n
		}

		return totalBytesParsed, nil
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
		r.parserState = Done
		return bytesParsed, nil
	}

	if bytesParsed == 0 {
		return 0, nil
	}

	return bytesParsed, nil
}
