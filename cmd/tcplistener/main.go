package main

import (
	"log"
	"net"
	"fmt"
	"github.com/TJ-R/httpfromtcp/internal/request"
)

func main() {
	l, err := net.Listen("tcp", ":42069")

	if err != nil {
		log.Fatal("Error with listening on port")
	}

	defer l.Close()
	for {
		con, err := l.Accept()
		defer con.Close()
		
		if err != nil {
			log.Printf("Error: %v\n", err)
		}

		log.Print("Connection Accepted")

		parsedRequest, err := request.RequestFromReader(con)
		if err != nil {
			log.Fatalf("Error when parsing request: %v\n", err)
		}

		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %v\n", parsedRequest.RequestLine.Method)
		fmt.Printf("- Target: %v\n", parsedRequest.RequestLine.RequestTarget)
		fmt.Printf("- Version: %v\n", parsedRequest.RequestLine.HttpVersion)

		log.Print("Connection Closed")
	}
}
