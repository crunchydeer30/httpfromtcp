package headers

import (
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
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
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
	assert.Equal(t, "application/json", headers.Get("Content-Type"))
	assert.Equal(t, 32, n)
	assert.False(t, done)

	// Test valid single header with extra whitespace
	headers = NewHeaders()
	data = []byte("   Content-Type:    application/json  \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "application/json", headers.Get("Content-Type"))
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
	require.NoError(t, err)
	assert.Equal(t, "application/json", headers.Get("Content-Type"))
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.True(t, done)

	// Invalid spacing header
	headers = NewHeaders()
	data = []byte("   Content-Type :    application/json  \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
}

func TestHeaders_InvalidCharacterInKey(t *testing.T) {
	headers := NewHeaders()
	data := []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)
	assert.Equal(t, "", headers.Get("Host"))
}

func TestHeaders_MultipleValuesForSameHeader(t *testing.T) {
	headers := NewHeaders()
	data := []byte("key: val1\r\nkey: val2\r\nkey: val3\r\n\r\n")

	readUntil := 0
	var (
		n    int
		done bool
		err  error
	)
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

	require.NoError(t, err)
	assert.Equal(t, "val1, val2, val3", headers.Get("key"))
	assert.True(t, done)
}
