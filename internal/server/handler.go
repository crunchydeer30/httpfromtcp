package server

import (
	"io"

	"github.com/crunchydeer30/httpfromtcp/internal/request"
	"github.com/crunchydeer30/httpfromtcp/internal/response"
)

type Handler func(w io.Writer, req *request.Request) *HandlerError

type HandlerError struct {
	StatusCode response.StatusCode
	Message    string
}

func WriteHandlerError(w io.Writer, handlerError *HandlerError) error {
	response.WriteStatusLine(w, handlerError.StatusCode)
	response.WriteHeaders(w, response.GetDefaultHeaders(len(handlerError.Message)))
	_, err := w.Write([]byte(handlerError.Message))
	return err
}
