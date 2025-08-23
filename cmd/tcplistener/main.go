package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"net"
)

func main() {
	tcpListener, err := net.Listen("tcp", "127.0.0.1:42069")
	if err != nil {
		return
	}
	defer func() {
		if err := tcpListener.Close(); err != nil {
			fmt.Printf("Error closing listener: %v\n", err)
		}
	}()
	for {
		tcpConn, err := tcpListener.Accept()
		if err != nil {
			return
		}
		fmt.Printf("a connection has been accepted\n")
		r, err := request.RequestFromReader(tcpConn)
		if err != nil {
			fmt.Printf("Error parsing request: %v\n", err)
			continue
		}
		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %v\n", r.RequestLine.Method)
		fmt.Printf("- Target: %v\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %v\n", r.RequestLine.HttpVersion)
	}
}
