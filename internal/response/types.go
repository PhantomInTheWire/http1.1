package response

import "io"

type WriterState int

const (
	StateInitial WriterState = iota
	StateStatusWritten
	StateHeadersWritten
	StateBodyWritten
)

type Writer struct {
	w     io.Writer
	state WriterState
}

type StatusCode int

const (
	StatusOK                  StatusCode = 200
	StatusBadRequest          StatusCode = 400
	StatusInternalServerError StatusCode = 500
)
