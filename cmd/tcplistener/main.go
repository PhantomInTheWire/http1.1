package main

import (
	"fmt"
	"httpfromtcp/internal/request"
	"io"
	"net"
	"strings"
)

func fRead(f io.ReadCloser, stream chan string) {
	defer close(stream)
	defer f.Close()
	currentLineContents := ""
	for {
		buffer := make([]byte, 8)
		n, err := f.Read(buffer)
		if err != nil {
			if err == io.EOF {
				if currentLineContents != "" {
					stream <- currentLineContents
				}
			}
			break
		}
		// copy everything inside the string
		str := string(buffer[:n])
		// split at new line
		parts := strings.Split(str, "\n")
		// print everything
		for i := 0; i < len(parts)-1; i++ {
			msg := fmt.Sprintf("%s%s", currentLineContents, parts[i])
			stream <- msg
			currentLineContents = ""
		}
		// add the last part for next iter since it might hot have \n
		currentLineContents += parts[len(parts)-1]
	}
}

func main() {
	tcpListener, err := net.Listen("tcp", "127.0.0.1:42069")
	if err != nil {
		return
	}
	defer tcpListener.Close()
	for {
		tcpConn, err := tcpListener.Accept()
		if err != nil {
			return
		}
		fmt.Printf("a connection has been accepted\n")
		r, err := request.RequestFromReader(tcpConn)
		fmt.Printf("Request line:\n")
		fmt.Printf("- Method: %v\n", r.RequestLine.Method)
		fmt.Printf("- Target: %v\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %v\n", r.RequestLine.HttpVersion)
	}
}
