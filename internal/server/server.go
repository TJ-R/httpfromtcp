package server

import (
	"fmt"
	"log"
	"net"

	"github.com/TJ-R/httpfromtcp/internal/request"
	"github.com/TJ-R/httpfromtcp/internal/response"
)

type Server struct {
	state ServerState
	listener net.Listener
	handler Handler
}

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

type ServerState int

type Handler func(w *response.Writer, req *request.Request)

const (
	Initalized ServerState = iota
	Closed
)

func Serve(port int, handler Handler) (*Server, error) {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err 
	}

	server := &Server {
		state: Initalized,
		listener: l,
		handler: handler,
	}

	go server.listen(handler)	
	return server, nil
}

func (s *Server) Close() error {
	s.state = Closed
	if s.listener != nil {
		return s.listener.Close()
	}

	return nil
}

func (s *Server) listen(handlerFunc Handler) {
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

	req, err := request.RequestFromReader(conn)
	if err != nil {
		log.Println(err)
	}

	writer := &response.Writer {
		W: conn,
	}

	s.handler(writer, req)
} 
