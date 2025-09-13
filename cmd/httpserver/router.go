package main

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"strings"
)

func router(w *response.Writer, req *request.Request) {
	path := req.RequestLine.RequestTarget

	switch {
	case path == "/":
		_ = w.WriteStatusLine(response.StatusOK)
		h := response.GetDefaultHeaders(len("Welcome to server 42069\n"))
		_ = w.WriteHeaders(h)
		_, _ = w.WriteBody([]byte("Welcome to server 42069\n"))
	case path == "/video":
		videoHandler(w, req)
	case strings.HasPrefix(path, "/httpbin/"):
		proxyHandler(w, req)
	default:
		_ = w.WriteStatusLine(response.StatusBadRequest)
		h := response.GetDefaultHeaders(len("Not Found"))
		_ = w.WriteHeaders(h)
		_, _ = w.WriteBody([]byte("Not Found"))
	}
}
