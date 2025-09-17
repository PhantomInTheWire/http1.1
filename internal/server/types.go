package server

import (
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"net"
)

type ServerState int

const (
	OpenState ServerState = iota
	ClosedState
)

type Server struct {
	State    ServerState
	Port     int
	Listener net.Listener
	handler  Handler
}

type HandlerError struct {
	StatusCode   int
	ErrorMessage string
}

type Handler func(w *response.Writer, req *request.Request)
