package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeadersParse(t *testing.T) {
	t.Run("Valid single header", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("Host: localhost:42069\r\n\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		require.NotNil(t, headers)
		assert.Equal(t, "localhost:42069", headers["Host"])
		assert.Equal(t, len("Host: localhost:42069\r\n\r\n"), n)
		assert.True(t, done)
	})

	t.Run("Valid single header with extra whitespace", func(t *testing.T) {
		headers := NewHeaders()
		data := []byte("    Host:    localhost:42069    \r\n\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		require.NotNil(t, headers)
		assert.Equal(t, "localhost:42069", headers["Host"]) // value should be trimmed
		assert.Equal(t, len("    Host:    localhost:42069    \r\n\r\n"), n)
		assert.True(t, done)
	})

	t.Run("Valid 2 headers with existing headers", func(t *testing.T) {
		headers := NewHeaders()
		headers["X-PreSet"] = "foo"

		data := []byte("Host: localhost:42069\r\nUser-Agent: TestClient\r\n\r\n")
		n, done, err := headers.Parse(data)
		require.NoError(t, err)
		require.NotNil(t, headers)

		// Verify both headers parsed + preset remains
		assert.Equal(t, "localhost:42069", headers["Host"])
		assert.Equal(t, "TestClient", headers["User-Agent"])
		assert.Equal(t, "foo", headers["X-PreSet"])
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
}
