# httpfromtcp

A RFC-compliant Go implementation of an HTTP/1.1 server using raw TCP sockets, without relying on Go's net/http package. This project demonstrates low-level HTTP protocol handling, including request parsing, header processing, response writing, and advanced features like chunked transfer encoding and trailers.

## Features

- **Custom HTTP Request Parser**: Parses HTTP/1.1 request lines, headers, and bodies from TCP connections, supporting Content-Length for bodies.
- **Header Management**: Validates and processes HTTP headers per RFC 7230, with case-insensitive keys and comma-separated values.
- **Response Writer**: Generates HTTP responses with status lines, headers, and bodies. Supports chunked encoding for streaming responses and trailers (e.g., SHA256 hash and content length).
- **Simple Routing**: Handles specific paths like `/video` (serves a static MP4 file) and `/httpbin/*` (proxies requests to httpbin.org with chunked responses).
- **TCP Server**: Non-blocking server with connection timeouts and graceful shutdown.
- **Utilities**: Includes a TCP listener for testing request parsing and a UDP sender (for unrelated testing or demo purposes).

## Directory Structure

```
httpfromtcp/
├── cmd/
│   ├── httpserver/     # Main HTTP server binary
│   │   └── main.go     # Entry point: starts server on port 42069
│   ├── tcplistener/    # Simple TCP listener to parse and print requests
│   │   └── main.go
│   └── udpsender/      # UDP client to send stdin data to localhost:42069
│       └── main.go
├── internal/
│   ├── headers/        # HTTP header parsing and validation
│   │   ├── headers.go
│   │   ├── types.go
│   │   ├── utils.go
│   │   └── headers_test.go
│   ├── request/        # HTTP request parsing logic
│   │   ├── constants.go
│   │   ├── parsing.go
│   │   ├── request.go
│   │   ├── request_test.go
│   │   ├── types.go
│   │   └── utils.go
│   ├── response/       # HTTP response writing (status, headers, body, chunked)
│   │   └── response.go
│   └── server/         # Core TCP server implementation
│       └── server.go
├── .github/workflows/ci.yml  # GitHub Actions CI
├── .gitignore
├── .pre-commit-config.yaml
├── go.mod               # Go modules (Go 1.24.6, testify for testing)
├── go.sum
├── README.md
└── messages.txt         # (Unused or placeholder file)
```

## Building and Running

### Prerequisites
- Go 1.24.6 or later

### Build
```bash
go mod tidy
go build ./cmd/httpserver
```

### Run the HTTP Server
```bash
./httpserver
```
The server listens on `localhost:42069`. It supports graceful shutdown via SIGINT/SIGTERM.

### Test Request Parsing
Build and run the TCP listener:
```bash
go build ./cmd/tcplistener
./tcplistener
```
Connect via `telnet localhost 42069` or `curl` to send requests and see parsed output.

### UDP Sender (Demo Utility)
```bash
go build ./cmd/udpsender
./udpsender
```
Type input to send UDP packets to `127.0.0.1:42069`.

## Endpoints

- **GET /video**: Serves `assets/vim.mp4` with Content-Type `video/mp4` and Connection: close.
- **GET /httpbin/***: Proxies to `https://httpbin.org` (e.g., `/httpbin/ip` fetches IP info). Uses chunked transfer encoding, streams response body, and adds trailers with SHA256 hash (`X-Content-SHA256`) and length (`X-Content-Length`).
- **Other paths**: Returns 400 Bad Request or 404-like response.

Example curl:
```bash
curl http://localhost:42069/video  # Streams video
curl http://localhost:42069/httpbin/user-agent  # Proxied response with trailers
```

## Components

### Request Parsing (`internal/request`)
- **State Machine**: Parses in phases (Pending → Headers → Body → Done).
- **RequestLine**: Extracts Method, Request-Target, HTTP-Version (only 1.1 supported).
- **Headers**: Integrated from `internal/headers`.
- **Body**: Accumulates based on Content-Length header.
- Handles partial reads with buffering.

### Headers (`internal/headers`)
- **Validation**: Ensures no space before colon, valid ASCII characters (32-126, no colon in keys).
- **Parsing**: Splits lines, trims spaces, lowercases keys, appends comma-separated values.
- **Methods**: `Parse` for incremental parsing, `Get/Set` for access.

### Server (`internal/server`)
- **Serve(port, handler)**: Starts TCP listener on `localhost:port`, accepts connections in goroutines.
- **Handler**: Function signature `func(*response.Writer, *request.Request)`.
- **Timeouts**: 30-second deadline per connection.
- **State**: Tracks Open/Closed.

### Response (`internal/response`)
- **Writer State Machine**: Ensures order (Status → Headers → Body/Trailers).
- **Status Codes**: 200 OK, 400 Bad Request, 500 Internal Server Error.
- **Chunked Encoding**: `WriteChunk` for streaming, `WriteChunkedBodyDone` for termination, `WriteTrailers` for metadata.
- **Defaults**: Helpers for Content-Length, Connection: close, text/plain.

## Testing

- Unit tests in `internal/headers/headers_test.go` and `internal/request/request_test.go` using `testify`.
- Run with:
  ```bash
  go test ./internal/...
  ```
- CI via GitHub Actions (`ci.yml`) for linting and testing.

