package main

import (
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/crunchydeer30/httpfromtcp/internal/request"
)

var ErrSomethingWentWrong = errors.New("something went wrong")

func main() {
	port := 42069
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("failed to create tcp listener on port %d: %s", port, err)
	}
	defer listener.Close()

	log.Printf("Listening for tcp connections on port %d\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatalf("failed to accept connection: %s", err)
		}
		log.Println("Received connection:", conn.RemoteAddr())

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	r, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Println("error reading request:", err)
		return
	}

	fmt.Printf("Request line:\n- Method: %s\n- Target: %s\n- Version: %s\n",
		r.RequestLine.Method,
		r.RequestLine.RequestTarget,
		r.RequestLine.HttpVersion)

	log.Println("Connection closed:", conn.RemoteAddr())
}
