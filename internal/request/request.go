package request

import (
	"io"
	"log"
	"strings"
	"fmt"
)

type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	requestStr, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	requestLine, err := parseRequestLine(string(requestStr))
	if err != nil {
		return nil, err
	}

	request := Request {
		RequestLine: *requestLine,
	}
	
	return &request, nil
}

func parseRequestLine(line string) (*RequestLine, error) {
	lines := strings.Split(line, "\r\n")
	requestLineStr := lines[0]
	requestSplit := strings.Split(requestLineStr, " ")

	log.Println(requestSplit)
	
	method := requestSplit[0]
	for _, c := range method {
		if c < 'A' || c > 'Z' {
			return nil, fmt.Errorf("Invalid method: %s", method)
		}
	}


	httpVersion := strings.Split(requestSplit[2], "/")
	if httpVersion[0] != "HTTP" {
		return nil, fmt.Errorf("Http Version is incorrect %s", httpVersion[0])
	}
	if httpVersion[1] != "1.1" {
		return nil, fmt.Errorf("Http Version is incorrect %s", httpVersion[1])
	}

	requestLine := RequestLine {
		HttpVersion: httpVersion[1],
		RequestTarget: requestSplit[1],
		Method: requestSplit[0],
	}

	return &requestLine, nil
}
