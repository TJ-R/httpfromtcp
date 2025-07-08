package main

import (
	"log"
	"net"
	"fmt"
	"errors"
	"io"
	"strings"
)

func main() {
	l, err := net.Listen("tcp", ":42069")

	if err != nil {
		log.Fatal("Error with listening on port")
	}

	defer l.Close()
	for {
		con, err := l.Accept()
		
		if err != nil {
			log.Printf("Error: %v\n", err)
		}

		log.Print("Connection Accepted")

		for line := range getLinesChannel(con) {
			fmt.Println(line)
		}

		con.Close()

		log.Print("Connection Closed")
	}
}

func getLinesChannel(f net.Conn) <- chan string {
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
