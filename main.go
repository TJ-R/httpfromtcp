package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	file, err := os.Open("messages.txt")
	if err != nil {
		fmt.Println("Error when opening messages.txt")
		return
	}

	for line := range getLinesChannel(file) {
		fmt.Printf("read: %v\n", line)
	}

}

func getLinesChannel(f io.ReadCloser) <- chan string {
	ch := make(chan string)

	go func() {
		defer f.Close()
		defer close(ch)
		currentLine := ""
		for {
			buf := make([]byte, 8)
			n, err := f.Read(buf) 

			if err != nil {
				if currentLine != "" {
					ch <- currentLine
					currentLine = ""
				}

				if errors.Is(err, io.EOF) {
					break
				}

				fmt.Println(err)
				return
			}

			stringArr := strings.Split(string(buf[:n]), "\n")
			for i := range len(stringArr)-1 {
				ch <- currentLine + stringArr[i]
				currentLine = ""
			}

			currentLine += stringArr[len(stringArr)-1]
		}
	}()

	return ch
}
