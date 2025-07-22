package server

import (
	"fmt"
	"net"
	"log"
	"github.com/TJ-R/httpfromtcp/internal/response"
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
	response.WriteStatusLine(conn, response.StatusOk)
	headers := response.GetDefaultHeaders(0)
	response.WriteHeaders(conn, headers)
}
