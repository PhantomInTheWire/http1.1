package main

import (
	"bufio"
	"log"
	"fmt"
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
	defer udpConn.Close()
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf(">")
		input, err := reader.ReadString('\n')
		if (err != nil) {
			log.Printf("error reading from stdin: %v", err)
			return
		}
		_, err = udpConn.Write([]byte(input))
		if (err != nil) {
			log.Printf("error writing to upd: %v", err)
			return
		}
	}
}
