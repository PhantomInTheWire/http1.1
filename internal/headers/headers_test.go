package headers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	t.Run("Valid single header", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("HOST: localhost:42069\r\n\r\n") // uppercase key
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		require.NotNil(t, headers)
		assert.Equal(t, "localhost:42069", headers["host"]) // key lowercased
		assert.Equal(t, len("HOST: localhost:42069\r\n\r\n"), n)
		assert.True(t, done)
	})

	t.Run("Valid single header with extra whitespace", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("    Host:    localhost:42069    \r\n\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		require.NotNil(t, headers)
		assert.Equal(t, "localhost:42069", headers["host"]) // value should be trimmed
		assert.Equal(t, len("    Host:    localhost:42069    \r\n\r\n"), n)
		assert.True(t, done)
	})

	t.Run("Valid 2 headers with existing headers", func(t *testing.T) {
		headers := NewHeaders()
		headers["x-preset"] = "foo"

		data := []byte("Host: localhost:42069\r\nUser-Agent: TestClient\r\n\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		require.NotNil(t, headers)

		// Verify both headers parsed + preset remains
		assert.Equal(t, "localhost:42069", headers["host"])
		assert.Equal(t, "TestClient", headers["user-agent"])
		assert.Equal(t, "foo", headers["x-preset"])
		assert.Equal(t, len("Host: localhost:42069\r\nUser-Agent: TestClient\r\n\r\n"), n)
		assert.True(t, done)
	})

	t.Run("Valid done", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("\r\n\r\n") // indicates empty headers, done
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		require.NotNil(t, headers)
		assert.Equal(t, 4, n) // len("\r\n\r\n")
		assert.True(t, done)
		assert.Empty(t, headers)
	})

	t.Run("Invalid spacing header", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("       Host : localhost:42069       \r\n\r\n")
		n, done, err := headers.Parse(data)
		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("Invalid character in header key", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("H©st: localhost:42069\r\n\r\n")
		n, done, err := headers.Parse(data)
		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("Mixed case headers", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Content-Type: application/json\r\nX-Custom-Header: Value123\r\n\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		require.NotNil(t, headers)
		assert.Equal(t, "application/json", headers["content-type"])
		assert.Equal(t, "Value123", headers["x-custom-header"])
		assert.Equal(t, len("Content-Type: application/json\r\nX-Custom-Header: Value123\r\n\r\n"), n)
		assert.True(t, done)
	})

	t.Run("Valid special characters in key", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("X-Test_Header: value\r\nX-Test-Header: value2\r\n\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		require.NotNil(t, headers)
		assert.Equal(t, "value", headers["x-test_header"])
		assert.Equal(t, "value2", headers["x-test-header"])
		assert.Equal(t, len("X-Test_Header: value\r\nX-Test-Header: value2\r\n\r\n"), n)
		assert.True(t, done)
	})

	t.Run("Invalid control character in key", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host\x00: localhost:42069\r\n\r\n")
		n, done, err := headers.Parse(data)
		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("Empty key", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte(": localhost:42069\r\n\r\n")
		n, done, err := headers.Parse(data)
		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})

	t.Run("Very long header", func(t *testing.T) {
		headers := NewHeaders()
		longValue := strings.Repeat("a", 1000)
		data := []byte("X-Long-Header: " + longValue + "\r\n\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		require.NotNil(t, headers)
		assert.Equal(t, longValue, headers["x-long-header"])
		assert.Equal(t, len("X-Long-Header: "+longValue+"\r\n\r\n"), n)
		assert.True(t, done)
	})

	t.Run("Multiple headers with mixed validity", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Valid-Header: value1\r\nInvalid©Key: value2\r\nAnother-Valid: value3\r\n\r\n")
		n, done, err := headers.Parse(data)
		require.Error(t, err)
		assert.Equal(t, 0, n)
		assert.False(t, done)
	})
}
