package main

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/crunchydeer30/httpfromtcp/internal/headers"
	"github.com/crunchydeer30/httpfromtcp/internal/request"
	"github.com/crunchydeer30/httpfromtcp/internal/response"
	"github.com/crunchydeer30/httpfromtcp/internal/server"
)

const port = 42069

func main() {
	server, err := server.Serve(port, streamHandler)
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

func streamHandler(w *response.ResponseWriter, req *request.Request) {
	res, err := http.Get("https://httpbin.org/stream/15")
	if err != nil {
		w.WriteStatusLine(response.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}
	defer res.Body.Close()

	w.WriteStatusLine(response.StatusOK)
	w.Headers.Set("Transfer-Encoding", "chunked")
	w.Headers.Set("content-type", "application/json")
	w.Headers.Set("Trailer", "X-Content-SHA256, X-Content-Length")
	w.WriteHeaders()

	buf := []byte{}
	for {
		data := make([]byte, 1024)
		read, err := res.Body.Read(data)
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		w.WriteChunkedBody(data[:read])
		buf = append(buf, data[:read]...)
	}
	w.WriteChunkedBodyDone()
	trailerHeaders := headers.NewHeaders()
	trailerHeaders.Set("X-Content-Length", strconv.Itoa(len(buf)))
	sum := sha256.Sum256(buf)
	trailerHeaders.Set("X-Content-SHA256", hex.EncodeToString(sum[:]))
	w.WriteTrailers(trailerHeaders)
}
