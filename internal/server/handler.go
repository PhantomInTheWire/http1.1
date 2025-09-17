package server

import (
	"fmt"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"net"
	"os"
	"time"
)

func (s *Server) handle(conn net.Conn) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "Handler panic recovered: %v\n", r)
		}
		if err := conn.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing connection: %v\n", err)
		}
	}()

	if err := conn.SetDeadline(time.Now().Add(30 * time.Second)); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting connection deadline: %v\n", err)
		return
	}

	r, err := request.RequestFromReader(conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing request: %v\n", err)
		return
	}

	w := response.NewWriter(conn)
	s.handler(w, r)
}
