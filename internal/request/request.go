package request

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/crunchydeer30/httpfromtcp/internal/headers"
)

var ErrMalformedRequestLine = fmt.Errorf("malformed request line")
var ErrUnsupportedHttpVersion = fmt.Errorf("unsupported http version")
var ErrUnsupportedHttpMethod = fmt.Errorf("unsupported http method")
var ErrIncompleteData = fmt.Errorf("incomplete data")

type parserStatus string

const (
	INITIALIZED     parserStatus = "initialized"
	DONE            parserStatus = "done"
	PARSING_HEADERS parserStatus = "parsing_headers"
)

const BUFFER_SIZE = 128
const CRLF = "\r\n"

type Request struct {
	RequestLine RequestLine
	State       parserStatus
	Headers     headers.Headers
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
		RequestLine: RequestLine{},
		State:       INITIALIZED,
		Headers:     headers.NewHeaders(),
	}

	buf := make([]byte, BUFFER_SIZE)
	eofReached := false

	for r.State != DONE {
		if readToIndex >= len(buf) {
			newBuf := make([]byte, len(buf)*2)
			copy(newBuf, buf[:readToIndex])
			buf = newBuf
		}

		n, err := reader.Read(buf[readToIndex:])
		if err != nil && err != io.EOF {
			return nil, err
		}
		if err == io.EOF {
			eofReached = true
		}
		readToIndex += n

		if eofReached && readToIndex == 0 {
			r.State = DONE
			break
		}

		parsed, err := r.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		if parsed == 0 && eofReached {
			return nil, ErrIncompleteData
		}

		if parsed > 0 {
			copy(buf, buf[parsed:readToIndex])
			readToIndex -= parsed
		}
	}

	return r, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.State != DONE {
		if totalBytesParsed >= len(data) {
			break
		}

		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return 0, err
		}
		if n == 0 {
			break
		}
		totalBytesParsed += n
	}

	return totalBytesParsed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.State {
	case INITIALIZED:
		n, err := r.parseRequestLine(data)
		if err != nil {
			return 0, err
		}
		if n == 0 {
			return n, nil
		}
		r.State = PARSING_HEADERS
		return n, nil
	case PARSING_HEADERS:
		n, done, err := r.Headers.Parse(data)
		if err != nil {
			return 0, err
		}
		if done {
			r.State = DONE
			return n, nil
		}
		return n, nil
	case DONE:
		return 0, errors.New("trying to read data in a done state")
	default:
		return 0, errors.New("unknown state")
	}
}

func (r *Request) parseRequestLine(data []byte) (int, error) {
	idx := strings.Index(string(data), CRLF)

	if idx == -1 {
		return 0, nil
	}

	requestLine := string(data[:idx])
	parts := strings.Split(requestLine, " ")

	if len(parts) != 3 {
		return 0, ErrMalformedRequestLine
	}

	r.RequestLine.HttpVersion = strings.TrimPrefix(parts[2], "HTTP/")
	r.RequestLine.RequestTarget = parts[1]
	r.RequestLine.Method = parts[0]

	if !r.RequestLine.ValidHttpVersion() {
		return 0, ErrUnsupportedHttpVersion
	}

	if !r.RequestLine.ValidMethod() {
		return 0, ErrUnsupportedHttpMethod
	}

	consumed := idx + len(CRLF)
	return consumed, nil
}
