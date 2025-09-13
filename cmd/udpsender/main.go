package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	host := "localhost"
	port := 42069

	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Fatalf("failed to resolve UDP addr %s:%d: %s", host, port, err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		log.Fatalf("failed to resolve dial up UDP connection %s:%d: %s", host, port, err)
	}
	defer conn.Close()

	log.Println("UDP connection established", conn)

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Println("Error reading input:", err)
			continue
		}
		_, err = conn.Write([]byte(line))
		if err != nil {
			log.Println("Error while writing to the UDP connection", conn.RemoteAddr(), err)
			continue
		}
	}
}
