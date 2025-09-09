package request

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
)

var ErrMalformedRequestLine = fmt.Errorf("malformed request line")
var ErrUnsupportedHttpVersion = fmt.Errorf("unsupported http version")
var ErrUnsupportedHttpMethod = fmt.Errorf("unsupported http method")
var separator = "\r\n"

type parserStatus string

const (
	INITIALIZED parserStatus = "initialized"
	DONE        parserStatus = "done"
)

const BUFFER_SIZE = 8
const NEW_LINE = "\r\n"

type Request struct {
	RequestLine RequestLine
	State       parserStatus
}

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

func (r *RequestLine) ValidMethod() bool {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	return slices.Contains(methods, r.Method)
}

func (r *RequestLine) ValidHttpVersion() bool {
	validVersions := []string{"1.1"}
	return slices.Contains(validVersions, r.HttpVersion)
}

func RequestFromReader(reader io.Reader) (*Request, error) {
	readToIndex := 0
	r := &Request{
		State: INITIALIZED,
	}

	buf := make([]byte, BUFFER_SIZE)
	for r.State != DONE {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf)
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		readToIndex += n

		if err == io.EOF {
			r.State = DONE
			break
		}
		if err != nil {
			return nil, err
		}

		parsed, err := r.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		copy(buf, buf[parsed:readToIndex])
		readToIndex -= parsed
	}

	return r, nil
}

func (r *Request) parse(data []byte) (int, error) {
	switch r.State {
	case INITIALIZED:
		n, err := r.parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return n, nil
		}
		r.State = DONE
		return n, nil
	case DONE:
		return 0, errors.New("trying to read data in a done state")
	default:
		return 0, errors.New("unknown state")
	}
}

func (r *Request) parseRequestLine(data []byte) (int, error) {

	idx := strings.Index(string(data), NEW_LINE)

	if idx == -1 {
		return 0, nil
	}

	requestLine := string(data[:idx])

	parts := strings.Split(requestLine, " ")

	if len(parts) != 3 {
		return 0, ErrMalformedRequestLine
	}

	rl := &RequestLine{
		HttpVersion:   strings.TrimPrefix(parts[2], "HTTP/"),
		RequestTarget: parts[1],
		Method:        parts[0],
	}

	if !rl.ValidHttpVersion() {
		return 0, ErrUnsupportedHttpVersion
	}

	if !rl.ValidMethod() {
		return 0, ErrUnsupportedHttpMethod
	}

	r.RequestLine = *rl
	consumed := idx + len(NEW_LINE)
	return consumed, nil
}
