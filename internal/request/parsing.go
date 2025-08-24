package request

import (
	"fmt"
	"strings"
)

// Constants for HTTP request parsing
const (
	// Request line component positions
	requestLineMethodIndex   = 0
	requestLineTargetIndex   = 1
	requestLineVersionIndex  = 2
	expectedRequestLineParts = 3

	// HTTP version parsing
	httpVersionPartsCount = 2
	httpVersionValueIndex = 1

	// Data extraction
	requestLineDataIndex = 0
	minDataPartsCount    = 1

	// String literals
	emptyString            = ""
	spaceDelimiter         = " "
	carriageReturnLineFeed = "\r\n"
	slashDelimiter         = "/"
)

func (r *Request) ParseRequestLine(data []byte) (int, error) {
	requestLine, consumed, err := parseRequestLine(string(data))
	if err != nil {
		return 0, err
	}
	if consumed == 0 {
		return 0, nil
	}
	if requestLine == nil {
		return 0, nil
	}

	r.RequestLine = *requestLine
	r.State = requestStateParsingHeaders
	return consumed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.State {
	case DoneState:
		return 0, nil
	case PendingState:
		return r.parsePendingState(data)
	case requestStateParsingHeaders:
		return r.parseHeadersState(data)
	default:
		return 0, ErrBadRequest
	}
}

func (r *Request) parsePendingState(data []byte) (int, error) {
	return r.ParseRequestLine(data)
}

func (r *Request) parseHeadersState(data []byte) (int, error) {
	n, done, err := r.Headers.Parse(data)
	if err != nil {
		return 0, err
	}
	if done {
		r.State = DoneState
	}
	return n, nil
}

func (r *Request) parse(data []byte) (int, error) {
	totalBytesParsed := 0
	for r.State != DoneState {
		n, err := r.parseSingle(data[totalBytesParsed:])
		if err != nil {
			return totalBytesParsed, err
		}
		totalBytesParsed += n
		if n == 0 {
			break
		}
	}
	return totalBytesParsed, nil
}

func parseRequestLine(RequestStr string) (*RequestLine, int, error) {
	requestLine, consumed, err := extractRequestLineFromData(RequestStr)
	if err != nil {
		return nil, 0, err
	}
	if consumed == 0 {
		return nil, 0, nil
	}

	rl, err := validateAndParseRequestLine(requestLine)
	if err != nil {
		return nil, 0, err
	}

	return rl, consumed, nil
}

func extractRequestLineFromData(data string) (string, int, error) {
	parts := strings.Split(data, carriageReturnLineFeed)
	if len(parts) == 1 && !strings.HasSuffix(data, carriageReturnLineFeed) {
		return emptyString, 0, nil
	}
	if len(parts) < minDataPartsCount {
		return emptyString, 0, ErrBadRequest
	}

	requestLine := parts[requestLineDataIndex]
	return requestLine, len(requestLine) + len(carriageReturnLineFeed), nil
}

func validateAndParseRequestLine(requestLine string) (*RequestLine, error) {
	if err := validateRequestLineFormat(requestLine); err != nil {
		return nil, err
	}

	return parseRequestLineComponents(requestLine)
}

func parseRequestLineComponents(requestLine string) (*RequestLine, error) {
	parts, err := splitRequestLine(requestLine)
	if err != nil {
		return nil, err
	}

	return createRequestLine(parts)
}

func splitRequestLine(requestLine string) ([]string, error) {
	parts := strings.Split(requestLine, spaceDelimiter)
	if len(parts) != expectedRequestLineParts {
		return nil, ErrBadRequest
	}
	return parts, nil
}

func createRequestLine(parts []string) (*RequestLine, error) {
	httpVersion, err := extractHttpVersion(parts[requestLineVersionIndex])
	if err != nil {
		return nil, err
	}

	return &RequestLine{
		HttpVersion:   httpVersion,
		RequestTarget: parts[requestLineTargetIndex],
		Method:        parts[requestLineMethodIndex],
	}, nil
}

func extractHttpVersion(versionPart string) (string, error) {
	parts := strings.Split(versionPart, slashDelimiter)
	if len(parts) != httpVersionPartsCount {
		return emptyString, ErrBadRequest
	}
	httpVersion := parts[httpVersionValueIndex]
	if !validateHttpVersion(httpVersion) {
		return emptyString, fmt.Errorf("%s: %s", ErrBadRequest.Error(), versionPart)
	}
	return httpVersion, nil
}
