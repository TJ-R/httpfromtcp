package headers

import (
	"bytes"
	"fmt"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers) 
}
	
func (h Headers) Parse(data []byte) (n int, done bool, err error) {

	index := bytes.Index(data, []byte("\r\n"))

	if index == -1 {
		return 0, false, nil
	}
	
	if index == 0 {
		return 0, true, nil
	}

	// Deal with stuff before \r\n
	headerParts := bytes.SplitN(data[:index], []byte(":"), 2) 

	fieldName := string(headerParts[0])

	// Any whitespace at end of field indicate space between name and :
	if fieldName != strings.TrimRight(fieldName, " ") {
		return 0, false, fmt.Errorf("Invalid spacing in header")
	}

	fieldName = strings.TrimSpace(fieldName)
	fieldValue := bytes.TrimSpace(headerParts[1])


	// Header valid
	h[fieldName] = string(fieldValue)
	return index + 2, false, nil
}
