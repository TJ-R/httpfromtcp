package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:42069")

	if err != nil {
		log.Fatal("Error with listening on port")
	}

	con, err := net.DialUDP("udp", nil, addr)
	defer con.Close()

	if err != nil {
		log.Fatal("Failed to Dial UDP")
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')

		if err != nil {
			log.Printf("Error when reading line: %v\n", err)
		}

		_, err = con.Write([]byte(line))

		if err != nil {
			log.Printf("Error when writing: %v\n", err)
		}
	}
}
