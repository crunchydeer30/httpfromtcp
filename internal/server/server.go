package server

import (
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/crunchydeer30/httpfromtcp/internal/request"
	"github.com/crunchydeer30/httpfromtcp/internal/response"
)

type Handler func(w *response.ResponseWriter, req *request.Request)

type Server struct {
	listener net.Listener
	handler  Handler
	closed   atomic.Bool
}

func Serve(port int, handler Handler) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: listener,
		handler:  handler,
		closed:   atomic.Bool{},
	}
	server.closed.Store(false)
	go server.listen()

	return server, nil
}

func (s *Server) Close() error {
	s.closed.Store(true)
	return s.listener.Close()
}

func (s *Server) listen() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.closed.Load() {
				return
			}
			log.Println("error accepting TCP connection", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	responseWriter := response.NewResponseWriter(conn)

	r, err := request.RequestFromReader(conn)
	if err != nil {
		log.Println("error reading request:", err)
		responseWriter.WriteStatusLine(response.StatusBadRequest)
		responseWriter.WriteHeaders()
		return
	}

	s.handler(responseWriter, r)
	responseWriter.Finalize()
}
