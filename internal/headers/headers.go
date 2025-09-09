package headers

import (
	"fmt"
	"strings"
)

type Headers map[string]string

var ErrMalformedHeader = fmt.Errorf("malformed header")

const CRLF = "\r\n"

func NewHeaders() Headers {
	return make(Headers)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := strings.Index(string(data), CRLF)

	if idx == -1 {
		return 0, false, nil
	}
	if idx == 0 {
		return 0, true, nil
	}

	headerLine := string(data[:idx])

	colonIdx := strings.Index(headerLine, ":")
	if idx == -1 {
		return 0, false, ErrMalformedHeader
	}

	key := strings.TrimLeft(headerLine[:colonIdx], " ")
	if strings.Contains(key, " ") {
		return 0, false, ErrMalformedHeader
	}

	value := strings.TrimSpace(headerLine[colonIdx+1:])

	h[key] = value

	return idx + len(CRLF), false, nil
}
