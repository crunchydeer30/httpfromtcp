package server

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"sync/atomic"

	"github.com/crunchydeer30/httpfromtcp/internal/request"
	"github.com/crunchydeer30/httpfromtcp/internal/response"
)

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
		if conn == nil {
			continue
		}
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

	r, err := request.RequestFromReader(conn)
	if err != nil {
		log.Println("error reading request:", err)
		handlerError := &HandlerError{
			StatusCode: response.BAD_REQUEST,
			Message:    err.Error(),
		}
		WriteHandlerError(conn, handlerError)
		return
	}

	handlerBuffer := bytes.NewBuffer([]byte{})

	if s.handler != nil {
		handlerError := s.handler(handlerBuffer, r)
		if handlerError != nil {
			WriteHandlerError(conn, handlerError)
			return
		}
	}

	headers := response.GetDefaultHeaders(handlerBuffer.Len())

	err = response.WriteStatusLine(conn, response.OK)
	if err != nil {
		log.Println("error writing status line:", err)
		return
	}

	err = response.WriteHeaders(conn, headers)
	if err != nil {
		log.Println("error writing headers:", err)
		return
	}

	err = response.WriteBody(conn, handlerBuffer.String())
	if err != nil {
		log.Println("error writing body:", err)
		return
	}
}
