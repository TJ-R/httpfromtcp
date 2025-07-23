package server

import (
	"bytes"
	"fmt"
	"io"
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

type Handler func(w io.Writer, req *request.Request) *HandlerError

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
		handleErr := HandlerError {
			StatusCode: response.StatusServerError,
			Message: fmt.Sprintf("%v", err),
		}
		handleErr.Write(conn)
	}

	buf := bytes.NewBuffer([]byte{})
	handleError := s.handler(buf, req)

	b := buf.Bytes()

	if handleError != nil {
		handleError.Write(conn)	
	}

	response.WriteStatusLine(conn, response.StatusOk)
	headers := response.GetDefaultHeaders(len(b))
	response.WriteHeaders(conn, headers)
	conn.Write(b)
}

func (he HandlerError) Write(w io.Writer) {
	response.WriteStatusLine(w, he.StatusCode)
	b := []byte(he.Message)
	headers := response.GetDefaultHeaders(len(b))
	response.WriteHeaders(w, headers)
	w.Write(b)
} 


