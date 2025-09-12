package headers

import (
	"fmt"
	"strings"
	"unicode"
)

const CRLF = "\r\n"

var ErrMalformedHeader = fmt.Errorf("malformed header")

type Headers map[string]string

func (h Headers) Get(key string) string {
	k := strings.ToLower(key)

	if value, ok := h[k]; ok {
		return value
	}
	return ""
}

func (h Headers) Set(key string, value string) {
	k := strings.ToLower(key)
	if _, ok := h[k]; ok {
		h[k] = h[k] + ", " + value
	} else {
		h[k] = value
	}
}

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
	if colonIdx == -1 {
		return 0, false, ErrMalformedHeader
	}

	key := strings.TrimLeft(headerLine[:colonIdx], " ")
	if !isValidHeader(key) {
		return 0, false, ErrMalformedHeader
	}

	value := strings.TrimSpace(headerLine[colonIdx+1:])

	h.Set(key, value)
	return idx + len(CRLF), false, nil
}

func isValidHeader(key string) bool {
	if strings.Contains(key, " ") {
		return false
	}

	for _, r := range key {
		switch {
		case unicode.IsUpper(r), unicode.IsLower(r), unicode.IsDigit(r):
			continue
		case strings.ContainsRune("!#$%&'*+-.^_`|~", r):
			continue
		default:
			return false
		}
	}
	return true

}
