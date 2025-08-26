package main

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"
)

const port = 42069

const badRequestHTML = `<html>
  <head>
    <title>400 Bad Request</title>
  </head>
  <body>
    <h1>Bad Request</h1>
    <p>Your request honestly kinda sucked.</p>
  </body>
</html>`

const internalErrorHTML = `<html>
  <head>
    <title>500 Internal Server Error</title>
  </head>
  <body>
    <h1>Internal Server Error</h1>
    <p>Okay, you know what? This one is on me.</p>
  </body>
</html>`

const successHTML = `<html>
  <head>
    <title>200 OK</title>
  </head>
  <body>
    <h1>Success!</h1>
    <p>Your request was an absolute banger.</p>
  </body>
</html>`

func myHandler(w *response.Writer, req *request.Request) {
	if req.RequestLine.RequestTarget == "/yourproblem" {
		_ = w.WriteStatusLine(response.StatusBadRequest)
		headers := response.GetDefaultHeaders(len(badRequestHTML))
		_ = headers.Set("Content-Type", "text/html")
		_ = w.WriteHeaders(headers)
		_, _ = w.WriteBody([]byte(badRequestHTML))
		return
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		_ = w.WriteStatusLine(response.StatusInternalServerError)
		headers := response.GetDefaultHeaders(len(internalErrorHTML))
		_ = headers.Set("Content-Type", "text/html")
		_ = w.WriteHeaders(headers)
		_, _ = w.WriteBody([]byte(internalErrorHTML))
		return
	}
	_ = w.WriteStatusLine(response.StatusOK)
	headers := response.GetDefaultHeaders(len(successHTML))
	_ = headers.Set("Content-Type", "text/html")
	_ = w.WriteHeaders(headers)
	_, _ = w.WriteBody([]byte(successHTML))
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
