package response

import (
	"io"
	"strconv"
	"strings"

	"github.com/crunchydeer30/httpfromtcp/internal/headers"
)

type StatusCode string

const (
	OK                    StatusCode = "200 OK"
	BAD_REQUEST           StatusCode = "400 Bad Request"
	INTERNAL_SERVER_ERROR StatusCode = "500 Internal Server Error"
)

func WriteStatusLine(w io.Writer, statusCode StatusCode) error {
	switch statusCode {
	case OK, BAD_REQUEST, INTERNAL_SERVER_ERROR:
		_, err := w.Write([]byte("HTTP/1.1 " + string(statusCode) + "\r\n"))
		if err != nil {
			return err
		}
		return nil
	default:
		return nil
	}
}

func GetDefaultHeaders(contentLen int) headers.Headers {
	headers := headers.NewHeaders()
	headers.Set("Content-Length", strconv.Itoa(contentLen))
	headers.Set("Connection", "close")
	headers.Set("Content-Type", "text/plain")
	return headers
}

func WriteHeaders(w io.Writer, headers headers.Headers) error {
	for k, v := range headers {
		values := strings.Split(v, ",")
		for _, val := range values {
			_, err := w.Write([]byte(k + ": " + val + "\r\n"))
			if err != nil {
				return err
			}
		}
	}
	_, err := w.Write([]byte("\r\n"))
	return err
}

func WriteBody(w io.Writer, body string) error {
	_, err := w.Write([]byte(body))
	return err
}
