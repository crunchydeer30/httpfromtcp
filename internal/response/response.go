package response

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/crunchydeer30/httpfromtcp/internal/headers"
)

type ResponseWriterState string

const CRLF = "\r\n"

const (
	WaitingForStatusLine ResponseWriterState = "WaitingForStatusLine"
	WaitingForHeaders    ResponseWriterState = "WaitingForHeaders"
	WaitingForBody       ResponseWriterState = "WaitingForBody"
)

type ResponseWriter struct {
	conn           net.Conn
	Headers        headers.Headers
	statusWritten  bool
	headersWritten bool
	bodyBuffer     *bytes.Buffer
}

func NewResponseWriter(conn net.Conn) *ResponseWriter {
	return &ResponseWriter{
		bodyBuffer:     bytes.NewBuffer([]byte{}),
		conn:           conn,
		statusWritten:  false,
		headersWritten: false,
		Headers:        headers.NewHeaders(),
	}
}

func (w *ResponseWriter) WriteStatusLine(statusCode StatusCode) error {
	if w.statusWritten {
		return fmt.Errorf("status line already written")
	}

	_, err := w.conn.Write([]byte("HTTP/1.1 " + statusCode.String() + " " + statusCode.StatusText() + "\r\n"))
	if err != nil {
		return err
	}

	w.statusWritten = true
	return nil
}

func (w *ResponseWriter) WriteHeaders() error {
	if w.headersWritten {
		return fmt.Errorf("headers already sent")
	}

	for k, v := range w.Headers {
		values := strings.Split(v, ",")
		for _, val := range values {
			_, err := w.conn.Write([]byte(k + ": " + val + "\r\n"))
			if err != nil {
				return err
			}
		}
	}
	_, err := w.conn.Write([]byte("\r\n"))
	if err != nil {
		return err
	}

	w.headersWritten = true
	return nil
}

func (w *ResponseWriter) Write(p []byte) error {
	w.bodyBuffer.Write(p)
	w.Headers.Replace("Content-Length", strconv.Itoa(w.bodyBuffer.Len()))
	return nil
}

func (w *ResponseWriter) WriteTrailers(h headers.Headers) {

}

func (w *ResponseWriter) Finalize() error {
	w.WriteStatusLine(200)
	w.SetDefaultHeaders(w.bodyBuffer.Len())
	w.WriteHeaders()
	_, err := w.conn.Write(w.bodyBuffer.Bytes())
	return err
}

func (w *ResponseWriter) SetDefaultHeaders(contentLen int) {
	if w.Headers.Get("Content-Length") == "" {
		w.Headers.Set("Content-Length", "0")
	}
	if w.Headers.Get("Connection") == "" {
		w.Headers.Set("Connection", "close")
	}
	if w.Headers.Get("Transfer-Encoding") == "chunked" {
		w.Headers.Delete("Content-Length")
		w.Headers.Delete("Connection")
	}
	if w.Headers.Get("Content-Type") == "" {
		w.Headers.Set("Content-Type", "text/plain")
	}
}

func (w *ResponseWriter) WriteChunkedBody(p []byte) (int, error) {
	n, _ := w.conn.Write([]byte(fmt.Sprintf("%x\r\n%s\r\n", len(p), p)))

	return n, nil
}

func (w *ResponseWriter) WriteChunkedBodyDone() (int, error) {
	w.conn.Write([]byte("0\r\n\r\n"))
	return 0, nil
}
