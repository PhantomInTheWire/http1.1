package main

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"httpfromtcp/internal/headers"
	"log"
	"os"
	"os/signal"
	"syscall"
	"strings"
	"io"
	"net/http"
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

func proxyHandler(w *response.Writer, req *request.Request) {
	if !strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin/") {
		_ = w.WriteStatusLine(response.StatusBadRequest)
		h := response.GetDefaultHeaders(len("Bad Request"))
		_ = w.WriteHeaders(h)
		_, _ = w.WriteBody([]byte("Bad Request"))
		return
	}

	path := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	targetURL := "https://httpbin.org" + path

	resp, err := http.Get(targetURL)
	if err != nil {
		_ = w.WriteStatusLine(response.StatusInternalServerError)
		h := response.GetDefaultHeaders(len("Internal Error"))
		_ = w.WriteHeaders(h)
		_, _ = w.WriteBody([]byte("Internal Error"))
		return
	}
	defer resp.Body.Close()

	// Write status line
	_ = w.WriteStatusLine(response.StatusCode(resp.StatusCode))

	// Copy headers except Content-Length
	h := headers.NewHeaders()
	for k, v := range resp.Header {
		if strings.ToLower(k) == "content-length" {
			continue
		}
		h[k] = strings.Join(v, ", ")
	}
	h["Transfer-Encoding"] = "chunked"
	_ = w.WriteHeaders(h)

	// Stream chunks
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			log.Printf("Read %d bytes from httpbin.org\n", n)
			_, _ = w.WriteChunk(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading body: %v", err)
			break
		}
	}

	// Final terminating chunk
	_ = w.WriteChunkedBodyDone()
}

func main() {
	server, err := server.Serve(port, proxyHandler)
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
