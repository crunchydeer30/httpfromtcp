package server

import (
	"fmt"
	"net"
	"sync/atomic"
)

type Server struct {
	listener net.Listener
	closed   atomic.Bool
}

func Serve(port int) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))

	if err != nil {
		return nil, err
	}

	server := &Server{
		listener: listener,
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
			fmt.Println("error accepting TCP connection", err)
			continue
		}
		go s.handle(conn)
	}
}

func (s *Server) handle(conn net.Conn) {
	defer conn.Close()

	conn.Write([]byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 13\r\n\r\nHello World!"))
}
