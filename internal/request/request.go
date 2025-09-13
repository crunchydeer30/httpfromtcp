package request

import (
	"errors"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/crunchydeer30/httpfromtcp/internal/headers"
)

var ErrMalformedRequestLine = fmt.Errorf("malformed request line")
var ErrUnsupportedHttpVersion = fmt.Errorf("unsupported http version")
var ErrUnsupportedHttpMethod = fmt.Errorf("unsupported http method")
var ErrIncompleteData = fmt.Errorf("incomplete data")
var ErrBodyTooLong = fmt.Errorf("body too long")
var ErrInvalidContentLength = fmt.Errorf("invalid content length")

type parserStatus string

const (
	INITIALIZED     parserStatus = "initialized"
	DONE            parserStatus = "done"
	PARSING_HEADERS parserStatus = "parsing_headers"
	PARSING_BODY    parserStatus = "parsing_body"
)

const BUFFER_SIZE = 128
const CRLF = "\r\n"
const CONTENT_LENGTH_HEADER = "content-length"

type Request struct {
	RequestLine RequestLine
	State       parserStatus
	Headers     headers.Headers
	Body        []byte
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

		parsed, err := r.parse(buf[:readToIndex])
		if err != nil {
			return nil, err
		}

		if parsed == 0 && eofReached && r.State != DONE {
			if r.State == PARSING_BODY {
				contentLength, err := r.getContentLength()
				if err != nil {
					return nil, err
				}
				if contentLength > len(r.Body) {
					return nil, ErrIncompleteData
				}
			} else {
				return nil, ErrIncompleteData
			}
		}
		if eofReached && readToIndex == 0 {
			r.State = DONE
			break
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
			contentLength, err := r.getContentLength()
			if err != nil {
				return 0, err
			}
			if contentLength == 0 {
				r.State = DONE
				return n, nil
			}
			r.State = PARSING_BODY
			return n, nil
		}
		return n, nil
	case PARSING_BODY:
		contentLength, err := r.getContentLength()
		if err != nil {
			return 0, err
		}
		if contentLength == 0 {
			r.State = DONE
			return 0, nil
		}

		remaining := contentLength - len(r.Body)
		toConsume := len(data)
		if toConsume > remaining {
			toConsume = remaining
		}

		r.Body = append(r.Body, data[:toConsume]...)

		if len(r.Body) > contentLength {
			return len(data), ErrBodyTooLong
		}
		if len(r.Body) == contentLength {
			r.State = DONE
			return len(data), nil
		}

		return len(data), nil
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

func (r Request) getContentLength() (int, error) {
	contentLengthHeader := r.Headers.Get(CONTENT_LENGTH_HEADER)
	if contentLengthHeader == "" {
		return 0, nil
	}
	contentLength, err := strconv.Atoi(contentLengthHeader)
	if err != nil || contentLength < 0 {
		return 0, ErrInvalidContentLength
	}
	return contentLength, nil
}
