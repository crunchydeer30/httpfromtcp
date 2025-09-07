package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
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

		linesChannel := getLinesChannel(conn)
		for line := range linesChannel {
			fmt.Printf("%s\n", line)
		}
		log.Println("Connection closed:", conn.RemoteAddr())
	}
}

func getLinesChannel(f io.ReadCloser) <-chan string {
	lines := make(chan string)

	go func() {
		defer close(lines)
		defer f.Close()

		currentLine := ""

		for {
			buf := make([]byte, 8)
			n, err := f.Read(buf)

			if n > 0 {
				parts := strings.Split(string(buf[:n]), "\n")
				for i := 0; i < len(parts)-1; i++ {
					lines <- currentLine + parts[i]
					currentLine = ""
				}
				currentLine += parts[len(parts)-1]
			}

			if err == io.EOF {
				if currentLine != "" {
					lines <- currentLine
				}
				break
			}

			if err != nil {
				break
			}
		}
	}()

	return lines
}
