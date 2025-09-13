package main

import (
	"crypto/sha256"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func videoHandler(w *response.Writer, req *request.Request) {
	wd, err := os.Getwd()
	if err != nil {
		_ = w.WriteStatusLine(response.StatusInternalServerError)
		h := response.GetDefaultHeaders(len("Video not found"))
		_ = w.WriteHeaders(h)
		_, _ = w.WriteBody([]byte("Video not found"))
		return
	}
	videoPath := filepath.Join(wd, "assets", "vim.mp4")

	videoData, err := os.ReadFile(videoPath)
	if err != nil {
		_ = w.WriteStatusLine(response.StatusInternalServerError)
		h := response.GetDefaultHeaders(len("Video not found"))
		_ = w.WriteHeaders(h)
		_, _ = w.WriteBody([]byte("Video not found"))
		return
	}

	_ = w.WriteStatusLine(response.StatusOK)

	h := headers.NewHeaders()
	h["Content-Length"] = fmt.Sprintf("%d", len(videoData))
	h["Content-Type"] = "video/mp4"
	h["Connection"] = "close"
	_ = w.WriteHeaders(h)

	_, _ = w.WriteBody(videoData)
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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	_ = w.WriteStatusLine(response.StatusCode(resp.StatusCode))

	h := headers.NewHeaders()
	for k, v := range resp.Header {
		if strings.ToLower(k) == "content-length" {
			continue
		}
		h[k] = strings.Join(v, ", ")
	}
	h["Transfer-Encoding"] = "chunked"
	h["Trailer"] = "X-Content-SHA256, X-Content-Length"
	_ = w.WriteHeaders(h)

	var bodyBuffer []byte
	buf := make([]byte, 1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			log.Printf("Read %d bytes from httpbin.org\n", n)
			bodyBuffer = append(bodyBuffer, buf[:n]...)
			_, _ = w.WriteChunk(buf[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Error reading body: %v\n", err)
			break
		}
	}

	_ = w.WriteChunkedBodyDone()

	hash := sha256.Sum256(bodyBuffer)
	hashHex := fmt.Sprintf("%x", hash)
	contentLength := len(bodyBuffer)

	trailers := headers.NewHeaders()
	trailers["X-Content-SHA256"] = hashHex
	trailers["X-Content-Length"] = fmt.Sprintf("%d", contentLength)
	_ = w.WriteTrailers(trailers)
}
