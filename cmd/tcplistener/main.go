package main

import (
	"fmt"
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

func getLinesChannel(f io.ReadCloser) <-chan string {
	stream := make(chan string)
	go fRead(f, stream)	
	return stream
}

func main() {
	tcpListener, err := net.Listen("tcp", "127.0.0.1:42069")
	if (err != nil) {
		return
	}
	defer tcpListener.Close()
	for {
		tcpConn, err := tcpListener.Accept()
		if (err != nil) {
			return
		}
		fmt.Printf("a connection has been accepted\n")
		stream := getLinesChannel(tcpConn)
		for line := range(stream) {
			fmt.Printf("%s\n", line)
		}
	}
}
