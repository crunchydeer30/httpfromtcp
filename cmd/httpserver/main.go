package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/crunchydeer30/httpfromtcp/internal/request"
	"github.com/crunchydeer30/httpfromtcp/internal/response"
	"github.com/crunchydeer30/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}

func handler(w *response.ResponseWriter, req *request.Request) {
	w.WriteStatusLine(response.StatusOK)
	w.Headers.Set("X-Test", "123")
	w.Headers.Set("Content-Type", "text/html")
	w.Write([]byte("<html><head><title>200 OK</title></head><body><h1>Success!</h1><p>Your request was an absolute banger.</p></body></html>"))
}
