package server

import (
	"fmt"
	"os"
)

func (s *Server) listen() {
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error accepting connection: %v\n", err)
			return
		}
		go func() {
			s.handle(conn)
		}()
	}
}
