package response

import (
	"bytes"
	"httpfromtcp/internal/headers"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriteStatusLine(t *testing.T) {
	tests := []struct {
		name       string
		statusCode StatusCode
		expected   string
	}{
		{"OK", StatusOK, "HTTP/1.1 200 OK\r\n"},
		{"Bad Request", StatusBadRequest, "HTTP/1.1 400 Bad Request\r\n"},
		{"Internal Server Error", StatusInternalServerError, "HTTP/1.1 500 Internal Server Error\r\n"},
		{"Unknown Status", StatusCode(404), "HTTP/1.1 404 Not Found\r\n"},
		{"Custom Status", StatusCode(999), "HTTP/1.1 999 Status\r\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := WriteStatusLine(&buf, tt.statusCode)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}

func TestGetDefaultHeaders(t *testing.T) {
	h := GetDefaultHeaders(42)
	assert.Equal(t, "42", h["Content-Length"])
	assert.Equal(t, "close", h["Connection"])
	assert.Equal(t, "text/plain", h["Content-Type"])
}

func TestWriteHeaders(t *testing.T) {
	var buf bytes.Buffer
	h := headers.NewHeaders()
	h["Content-Type"] = "application/json"
	h["X-Custom"] = "value"

	err := WriteHeaders(&buf, h)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "Content-Type: application/json")
	assert.Contains(t, output, "X-Custom: value")
	assert.Contains(t, output, "\r\n\r\n")
}

func TestWriteBody(t *testing.T) {
	var buf bytes.Buffer
	body := "Hello, World!"

	err := WriteBody(&buf, body)
	require.NoError(t, err)
	assert.Equal(t, body, buf.String())
}

func TestNewWriter(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	assert.NotNil(t, w)
	assert.Equal(t, StateInitial, w.state)
	assert.Equal(t, &buf, w.w)
}

func TestWriter_WriteStatusLine(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	// First write should succeed
	err := w.WriteStatusLine(StatusOK)
	require.NoError(t, err)
	assert.Equal(t, StateStatusWritten, w.state)
	assert.Equal(t, "HTTP/1.1 200 OK\r\n", buf.String())

	// Second write should fail
	err = w.WriteStatusLine(StatusBadRequest)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot write status line")
}

func TestWriter_WriteHeaders(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	// Should fail before status line
	h := GetDefaultHeaders(0)
	err := w.WriteHeaders(h)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status line not written yet")

	// Should succeed after status line
	require.NoError(t, w.WriteStatusLine(StatusOK))
	err = w.WriteHeaders(h)
	require.NoError(t, err)
	assert.Equal(t, StateHeadersWritten, w.state)
}

func TestWriter_WriteBody(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	// Should fail before headers
	_, err := w.WriteBody([]byte("test"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "headers not written yet")

	// Should succeed after headers
	require.NoError(t, w.WriteStatusLine(StatusOK))
	require.NoError(t, w.WriteHeaders(GetDefaultHeaders(4)))
	n, err := w.WriteBody([]byte("test"))
	require.NoError(t, err)
	assert.Equal(t, 4, n)
	assert.Equal(t, StateBodyWritten, w.state)
}

func TestWriter_WriteChunk(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	// Test empty chunk
	n, err := w.WriteChunk([]byte{})
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, "", buf.String())

	// Test normal chunk
	data := []byte("Hello")
	n, err = w.WriteChunk(data)
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	expected := "5\r\nHello\r\n"
	assert.Equal(t, expected, buf.String())
}

func TestWriter_WriteChunkedBodyDone(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	err := w.WriteChunkedBodyDone()
	require.NoError(t, err)
	assert.Equal(t, "0\r\n\r\n", buf.String())
	assert.Equal(t, StateBodyWritten, w.state)
}

func TestWriter_WriteTrailers(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	// Should fail before body is written
	h := headers.NewHeaders()
	h["X-Trailer"] = "value"
	err := w.WriteTrailers(h)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "body not written yet")

	// Should succeed after body is written
	require.NoError(t, w.WriteStatusLine(StatusOK))
	require.NoError(t, w.WriteHeaders(GetDefaultHeaders(0)))
	_, err = w.WriteBody([]byte{})
	require.NoError(t, err)
	err = w.WriteTrailers(h)
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "X-Trailer: value\r\n\r\n")
}

func TestWriter_StateTransitions(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	// Initial state
	assert.Equal(t, StateInitial, w.state)

	// After status line
	require.NoError(t, w.WriteStatusLine(StatusOK))
	assert.Equal(t, StateStatusWritten, w.state)

	// After headers
	require.NoError(t, w.WriteHeaders(GetDefaultHeaders(0)))
	assert.Equal(t, StateHeadersWritten, w.state)

	// After body
	_, err := w.WriteBody([]byte{})
	require.NoError(t, err)
	assert.Equal(t, StateBodyWritten, w.state)
}

func TestWriter_WriteChunkedResponse(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	// Write status and headers
	require.NoError(t, w.WriteStatusLine(StatusOK))
	h := headers.NewHeaders()
	h["Transfer-Encoding"] = "chunked"
	require.NoError(t, w.WriteHeaders(h))

	// Write chunks
	_, err := w.WriteChunk([]byte("Hello"))
	require.NoError(t, err)
	_, err = w.WriteChunk([]byte(" "))
	require.NoError(t, err)
	_, err = w.WriteChunk([]byte("World"))
	require.NoError(t, err)

	// Finish chunked body
	require.NoError(t, w.WriteChunkedBodyDone())

	// Write trailers
	trailers := headers.NewHeaders()
	trailers["X-Checksum"] = "abc123"
	require.NoError(t, w.WriteTrailers(trailers))

	expected := "HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\n" +
		"5\r\nHello\r\n1\r\n \r\n5\r\nWorld\r\n0\r\n\r\nX-Checksum: abc123\r\n\r\n"
	assert.Equal(t, expected, buf.String())
}

func TestWriter_WriteFullResponse(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	// Write complete response
	require.NoError(t, w.WriteStatusLine(StatusOK))
	require.NoError(t, w.WriteHeaders(GetDefaultHeaders(13)))
	_, err := w.WriteBody([]byte("Hello, World!"))
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "HTTP/1.1 200 OK")
	assert.Contains(t, output, "Content-Length: 13")
	assert.Contains(t, output, "Connection: close")
	assert.Contains(t, output, "Content-Type: text/plain")
	assert.Contains(t, output, "\r\n\r\nHello, World!")
}

func TestWriteSimpleResponse(t *testing.T) {
	var buf bytes.Buffer

	err := WriteSimpleResponse(&buf, StatusOK, "text/html", "<h1>Hello</h1>")
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "HTTP/1.1 200 OK")
	assert.Contains(t, output, "Content-Length: 14")
	assert.Contains(t, output, "Connection: close")
	assert.Contains(t, output, "Content-Type: text/html")
	assert.Contains(t, output, "\r\n\r\n<h1>Hello</h1>")
}

func TestWriteSimpleResponse_DefaultContentType(t *testing.T) {
	var buf bytes.Buffer

	err := WriteSimpleResponse(&buf, StatusOK, "", "Hello World")
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "HTTP/1.1 200 OK")
	assert.Contains(t, output, "Content-Length: 11")
	assert.Contains(t, output, "Connection: close")
	assert.Contains(t, output, "Content-Type: text/plain")
	assert.Contains(t, output, "\r\n\r\nHello World")
}

func TestWriteChunkedResponse(t *testing.T) {
	var buf bytes.Buffer

	chunks := [][]byte{
		[]byte("Hello"),
		[]byte(" "),
		[]byte("World"),
	}

	trailers := headers.NewHeaders()
	trailers["X-Checksum"] = "abc123"

	err := WriteChunkedResponse(&buf, StatusOK, "text/plain", chunks, trailers)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "HTTP/1.1 200 OK")
	assert.Contains(t, output, "Transfer-Encoding: chunked")
	assert.Contains(t, output, "Content-Type: text/plain")
	assert.Contains(t, output, "Connection: close")
	assert.Contains(t, output, "\r\n\r\n5\r\nHello\r\n1\r\n \r\n5\r\nWorld\r\n0\r\n\r\nX-Checksum: abc123\r\n\r\n")
}

func TestWriteChunkedResponse_NoTrailers(t *testing.T) {
	var buf bytes.Buffer

	chunks := [][]byte{
		[]byte("Test"),
	}

	err := WriteChunkedResponse(&buf, StatusOK, "", chunks, nil)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "HTTP/1.1 200 OK")
	assert.Contains(t, output, "Transfer-Encoding: chunked")
	assert.Contains(t, output, "Connection: close")
	assert.Contains(t, output, "\r\n\r\n4\r\nTest\r\n0\r\n\r\n")
}

func TestWriteErrorResponse(t *testing.T) {
	var buf bytes.Buffer

	err := WriteErrorResponse(&buf, StatusBadRequest, "Invalid request")
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "HTTP/1.1 400 Bad Request")
	assert.Contains(t, output, "Content-Length: 15")
	assert.Contains(t, output, "Connection: close")
	assert.Contains(t, output, "Content-Type: text/plain")
	assert.Contains(t, output, "\r\n\r\nInvalid request")
}

func TestWriteJSONResponse(t *testing.T) {
	var buf bytes.Buffer

	jsonData := `{"message": "success", "code": 200}`
	err := WriteJSONResponse(&buf, StatusOK, jsonData)
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "HTTP/1.1 200 OK")
	assert.Contains(t, output, "Content-Length: 35")
	assert.Contains(t, output, "Connection: close")
	assert.Contains(t, output, "Content-Type: application/json")
	assert.Contains(t, output, "\r\n\r\n"+jsonData)
}
