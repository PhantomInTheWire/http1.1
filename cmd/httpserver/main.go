package main

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

func myHandler(w io.Writer, req *request.Request) *server.HandlerError {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		return &server.HandlerError{StatusCode: 400, ErrorMessage: "Your problem is not my problem\n"}
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		return &server.HandlerError{StatusCode: 500, ErrorMessage: "Woopsie, my bad\n"}
	}
	if _, err := w.Write([]byte("All good, frfr\n")); err != nil {
		return &server.HandlerError{StatusCode: 500, ErrorMessage: "Internal server error\n"}
	}
	return nil
}

func main() {
	server, err := server.Serve(port, myHandler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer func() {
		if err := server.Close(); err != nil {
			log.Printf("Error closing server: %v", err)
		}
	}()
	log.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Server gracefully stopped")
}
