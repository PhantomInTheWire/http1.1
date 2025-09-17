package server

import (
	"bytes"
	"errors"
	"net"
	"testing"
	"time"

	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockConn struct {
	readData        *bytes.Buffer
	writeData       *bytes.Buffer
	closed          bool
	deadline        time.Time
	closeFunc       func() error
	setDeadlineFunc func(time.Time) error
}

func newMockConn(data string) *mockConn {
	return &mockConn{
		readData:  bytes.NewBufferString(data),
		writeData: &bytes.Buffer{},
		closed:    false,
		closeFunc: func() error {
			return nil
		},
		setDeadlineFunc: func(t time.Time) error {
			return nil
		},
	}
}

func (m *mockConn) Read(b []byte) (n int, err error) {
	if m.closed {
		return 0, errors.New("connection closed")
	}
	return m.readData.Read(b)
}

func (m *mockConn) Write(b []byte) (n int, err error) {
	if m.closed {
		return 0, errors.New("connection closed")
	}
	return m.writeData.Write(b)
}

func (m *mockConn) Close() error {
	m.closed = true
	return m.closeFunc()
}

func (m *mockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}
}

func (m *mockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}
}

func (m *mockConn) SetDeadline(t time.Time) error {
	m.deadline = t
	return m.setDeadlineFunc(t)
}

func (m *mockConn) SetReadDeadline(t time.Time) error {
	return m.SetDeadline(t)
}

func (m *mockConn) SetWriteDeadline(t time.Time) error {
	return m.SetDeadline(t)
}

type mockListener struct {
	conns  chan net.Conn
	closed bool
}

func newMockListener() *mockListener {
	return &mockListener{
		conns:  make(chan net.Conn, 10),
		closed: false,
	}
}

func (m *mockListener) Accept() (net.Conn, error) {
	if m.closed {
		return nil, errors.New("listener closed")
	}
	conn, ok := <-m.conns
	if !ok {
		return nil, errors.New("listener closed")
	}
	return conn, nil
}

func (m *mockListener) Close() error {
	if m.closed {
		return errors.New("listener already closed")
	}
	m.closed = true
	close(m.conns)
	return nil
}

func (m *mockListener) Addr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}
}

func (m *mockListener) addConn(conn net.Conn) {
	m.conns <- conn
}

func TestServe_Success(t *testing.T) {
	server, err := Serve(0, func(w *response.Writer, req *request.Request) {})
	require.NoError(t, err)
	require.NotNil(t, server)
	assert.Equal(t, OpenState, server.State)
	assert.Equal(t, 0, server.Port)
	assert.NotNil(t, server.Listener)

	err = server.Close()
	require.NoError(t, err)
}

func TestServe_InvalidPort(t *testing.T) {
	_, err := Serve(-1, nil)
	assert.Error(t, err)
}

func TestServer_Close(t *testing.T) {
	server, err := Serve(0, func(w *response.Writer, req *request.Request) {})
	require.NoError(t, err)

	err = server.Close()
	require.NoError(t, err)
	assert.Equal(t, ClosedState, server.State)

	err = server.Close()
	assert.Error(t, err)
}

func TestHandle_ValidRequest(t *testing.T) {
	requestData := "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"
	conn := newMockConn(requestData)

	server := &Server{
		State: OpenState,
		handler: func(w *response.Writer, req *request.Request) {
			assert.Equal(t, "GET", req.RequestLine.Method)
			assert.Equal(t, "/", req.RequestLine.RequestTarget)
			assert.Equal(t, "1.1", req.RequestLine.HttpVersion)
			assert.Equal(t, "localhost", req.Headers["host"])

			err := w.WriteStatusLine(response.StatusOK)
			require.NoError(t, err)
			err = w.WriteHeaders(response.GetDefaultHeaders(0))
			require.NoError(t, err)
			_, err = w.WriteBody([]byte{})
			require.NoError(t, err)
		},
	}

	server.handle(conn)

	assert.True(t, conn.closed)

	assert.True(t, !conn.deadline.IsZero())

	responseData := conn.writeData.String()
	assert.Contains(t, responseData, "HTTP/1.1 200 OK")
}

func TestHandle_InvalidRequest(t *testing.T) {
	requestData := "INVALID REQUEST\r\n\r\n"
	conn := newMockConn(requestData)

	server := &Server{
		State: OpenState,
		handler: func(w *response.Writer, req *request.Request) {
			t.Fatal("Handler should not be called for invalid request")
		},
	}

	server.handle(conn)

	assert.True(t, conn.closed)
}

func TestHandle_HandlerError(t *testing.T) {
	requestData := "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"
	conn := newMockConn(requestData)

	server := &Server{
		State: OpenState,
		handler: func(w *response.Writer, req *request.Request) {
			panic("handler error")
		},
	}

	assert.NotPanics(t, func() {
		server.handle(conn)
	})

	assert.True(t, conn.closed)
}

func TestHandle_ConnectionCloseError(t *testing.T) {
	conn := &mockConn{
		readData:  bytes.NewBufferString("GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"),
		writeData: &bytes.Buffer{},
		closed:    false,
		closeFunc: func() error {
			return errors.New("close error")
		},
		setDeadlineFunc: func(t time.Time) error {
			return nil
		},
	}

	server := &Server{
		State:   OpenState,
		handler: func(w *response.Writer, req *request.Request) {},
	}

	assert.NotPanics(t, func() {
		server.handle(conn)
	})
}

func TestHandle_SetDeadlineError(t *testing.T) {
	conn := &mockConn{
		readData:  bytes.NewBufferString("GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"),
		writeData: &bytes.Buffer{},
		closed:    false,
		closeFunc: func() error {
			return nil
		},
		setDeadlineFunc: func(t time.Time) error {
			return errors.New("deadline error")
		},
	}

	server := &Server{
		State:   OpenState,
		handler: func(w *response.Writer, req *request.Request) {},
	}

	assert.NotPanics(t, func() {
		server.handle(conn)
	})

	assert.True(t, conn.closed)
}

func TestListen_AcceptConnections(t *testing.T) {
	listener := newMockListener()
	server := &Server{
		State:    OpenState,
		Listener: listener,
		handler:  func(w *response.Writer, req *request.Request) {},
	}

	go server.listen()

	conn := newMockConn("GET / HTTP/1.1\r\nHost: localhost\r\n\r\n")
	listener.addConn(conn)

	time.Sleep(10 * time.Millisecond)

	_ = listener.Close()

	assert.True(t, conn.closed)
}

func TestListen_ListenerError(t *testing.T) {
	listener := newMockListener()
	server := &Server{
		State:    OpenState,
		Listener: listener,
		handler:  func(w *response.Writer, req *request.Request) {},
	}

	_ = listener.Close()

	assert.NotPanics(t, func() {
		server.listen()
	})
}

func TestServerState(t *testing.T) {
	assert.Equal(t, 0, int(OpenState))
	assert.Equal(t, 1, int(ClosedState))
}

func TestHandlerError(t *testing.T) {
	err := HandlerError{
		StatusCode:   500,
		ErrorMessage: "Internal Server Error",
	}
	assert.Equal(t, 500, err.StatusCode)
	assert.Equal(t, "Internal Server Error", err.ErrorMessage)
}

func TestHandler_Type(t *testing.T) {
	var h Handler = func(w *response.Writer, req *request.Request) {}
	assert.NotNil(t, h)
}

func BenchmarkHandle(b *testing.B) {
	requestData := "GET / HTTP/1.1\r\nHost: localhost\r\n\r\n"
	server := &Server{
		State: OpenState,
		handler: func(w *response.Writer, req *request.Request) {
			_ = w.WriteStatusLine(response.StatusOK)
			_ = w.WriteHeaders(response.GetDefaultHeaders(0))
			_, _ = w.WriteBody([]byte{})
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn := newMockConn(requestData)
		server.handle(conn)
	}
}

func BenchmarkServe(b *testing.B) {
	for i := 0; i < b.N; i++ {
		server, err := Serve(0, func(w *response.Writer, req *request.Request) {})
		if err != nil {
			b.Fatal(err)
		}
		_ = server.Close()
	}
}
