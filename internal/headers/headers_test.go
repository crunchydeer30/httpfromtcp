package headers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParser(t *testing.T) {
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.Equal(t, 23, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test valid single header
	headers = NewHeaders()
	data = []byte("Content-Type: application/json\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "application/json", headers["Content-Type"])
	assert.Equal(t, 32, n)
	assert.False(t, done)

	// Test valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("   Content-Type:    application/json  \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "application/json", headers["Content-Type"])
	assert.Equal(t, 40, n)
	assert.False(t, done)

	// Test 2 headers
	headers = NewHeaders()
	data = []byte("Content-Type: application/json\r\nHost: localhost:42069\r\n\r\n")
	n = 0
	readUntil := 0
	for {
		n, done, err = headers.Parse(data[readUntil:])
		readUntil += n
		if done == true {
			break
		}
		if err != nil {
			break
		}
	}
	fmt.Println("n", n, "done", done, "err", err)
	require.NoError(t, err)
	assert.Equal(t, "application/json", headers["Content-Type"])
	assert.Equal(t, "localhost:42069", headers["Host"])
	assert.True(t, done)

	// Invalid spacing header
	// Test valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("   Content-Type :    application/json  \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}
