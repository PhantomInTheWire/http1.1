package server

import (
	"fmt"
	"net"
	"strconv"
)

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
	s.State = ClosedState
	return err
}
