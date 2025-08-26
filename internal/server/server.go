package server

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"net"
	"os"
	"strconv"
)

type ServerState int

const (
	OpenState ServerState = iota
	ClosedState
)

type Server struct {
	State    ServerState
	Port     int
	Listener net.Listener
	handler  Handler
}

type HandlerError struct {
	StatusCode   int
	ErrorMessage string
}

func (h HandlerError) writeHandlerError(w io.Writer) {
	if err := response.WriteStatusLine(w, response.StatusCode(h.StatusCode)); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing status line: %v\n", err)
		return
	}
	body := []byte(h.ErrorMessage)
	headers := response.GetDefaultHeaders(len(body))
	if err := response.WriteHeaders(w, headers); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing headers: %v\n", err)
		return
	}
	if _, err := w.Write(body); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing body: %v\n", err)
		return
	}
}

type Handler func(w io.Writer, req *request.Request) *HandlerError

func Serve(port int, h Handler) (*Server, error) {
	portString := strconv.Itoa(port)
	tcpListener, err := net.Listen("tcp", "localhost:"+portString)
	if err != nil {
		return nil, err
	}
	s := Server{
		State:    OpenState,
		Port:     port,
		Listener: tcpListener,
		handler:  h,
	}
	go s.listen()
	return &s, nil
}

func (s *Server) Close() error {
	fmt.Printf("closing the server on port: %v\n", s.Port)
	err := s.Listener.Close()
	return err
}

func (s *Server) listen() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			fmt.Fprintln(os.Stderr, fmt.Errorf("error in connection %v", err))
		}
		go func() {
			s.handle(conn)
		}()
	}
}

func (s *Server) handle(conn net.Conn) {
	// request parsing
	r, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing request: %v\n", err)
		return
	}
	var buf bytes.Buffer

	hErr := s.handler(&buf, r)
	if hErr != nil {
		hErr.writeHandlerError(conn)
		return
	}
	// resposnse stuff
	if err := response.WriteStatusLine(conn, response.StatusOK); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing status line: %v\n", err)
		return
	}
	body := buf.String()
	headers := response.GetDefaultHeaders(len([]byte(body)))
	if err := response.WriteHeaders(conn, headers); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing headers: %v\n", err)
		return
	}
	if err := response.WriteBody(conn, body); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing body: %v\n", err)
		return
	}
	if err := conn.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Error closing connection: %v\n", err)
	}
}
