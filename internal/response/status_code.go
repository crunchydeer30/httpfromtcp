package response

import "strconv"

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)

func (s StatusCode) String() string {
	return strconv.Itoa(int(s))
}

func (s StatusCode) StatusText() string {
	switch s {
	case StatusOK:
		return "OK"
	case StatusBadRequest:
		return "Bad Request"
	case StatusInternalServerError:
		return "Internal Server Error"
	default:
		return ""
	}
}
