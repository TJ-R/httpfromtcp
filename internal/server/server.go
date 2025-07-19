package server

import (
	"fmt"
	"net"
	"log"
)

type Server struct {
	state ServerState
	listener net.Listener
}

type ServerState int

const (
	Initalized ServerState = iota
	Closed
)

func Serve(port int) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err 
	}

	server := &Server {
		state: Initalized,
		listener: l,
	}

	go server.listen()	
	return server, nil
}

func (s *Server) Close() error {
	s.state = Closed
	if s.listener != nil {
		return s.listener.Close()
	}

	return nil
}

func (s *Server) listen() {
	for s.state != Closed {
		conn, err := s.listener.Accept()	
		if err != nil {
			if s.state == Closed {
				return
			}
			log.Printf("Error accepting connection: %v", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()
	conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\n\r\nHello World!\n"))
}
