package server

import (
	"fmt"
	"httpfromtcp/internal/response"
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
}

func Serve(port int) (*Server, error) {
	portString := strconv.Itoa(port)
	tcpListener, err := net.Listen("tcp", "localhost:"+portString)
	if err != nil {
		return nil, err
	}
	s := Server{
		State:    OpenState,
		Port:     port,
		Listener: tcpListener,
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
	if err := response.WriteStatusLine(conn, response.StatusOK); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing status line: %v\n", err)
		return
	}
	h := response.GetDefaultHeaders(0)
	if err := response.WriteHeaders(conn, h); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing headers: %v\n", err)
		return
	}
	if err := conn.Close(); err != nil {
		fmt.Fprintf(os.Stderr, "Error closing connection: %v\n", err)
	}
}
