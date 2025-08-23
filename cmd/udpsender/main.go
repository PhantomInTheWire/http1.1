package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:42069")
	if err != nil {
		return
	}
	udpConn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		return
	}
	defer func() {
		if err := udpConn.Close(); err != nil {
			fmt.Printf("Error closing UDP connection: %v\n", err)
		}
	}()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf(">")
		input, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("error reading from stdin: %v", err)
			return
		}
		_, err = udpConn.Write([]byte(input))
		if err != nil {
			log.Printf("error writing to upd: %v", err)
			return
		}
	}
}
