package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"net"
	"os"
	"strconv"
	"time"
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

type Handler func(w *response.Writer, req *request.Request)

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
			fmt.Fprintf(os.Stderr, "Error accepting connection: %v\n", err)
		}
		go func() {
			s.handle(conn)
		}()
	}
}

func (s *Server) handle(conn net.Conn) {
	defer func() {
		if err := conn.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing connection: %v\n", err)
		}
	}()

	// Set connection deadline to prevent hanging
	if err := conn.SetDeadline(time.Now().Add(30 * time.Second)); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting connection deadline: %v\n", err)
		return
	}

	// request parsing
	r, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing request: %v\n", err)
		return
	}

	w := response.NewWriter(conn)
	s.handler(w, r)
}
