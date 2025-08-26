package request

import (
	"fmt"
	"io"
	"strconv"
	"strings"
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
	r.State = ParsingHeadersState
	return consumed, nil
}

func (r *Request) parseSingle(data []byte) (int, error) {
	switch r.State {
	case DoneState:
		return 0, nil
	case PendingState:
		return r.parsePendingState(data)
	case ParsingHeadersState:
		return r.parseHeadersState(data)
	case ParsingBodyState:
		return r.parseBodyState(data)
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
		r.State = ParsingBodyState
	}
	return n, nil
}

func (r *Request) parseBodyState(data []byte) (int, error) {
	contentLengthString, err := r.Headers.Get("Content-Length")
	if err != nil {
		return 0, err
	}
	if contentLengthString == "" {
		// No Content-Length header means no body expected
		// If we have data, it might be connection artifacts, so we ignore it
		r.State = DoneState
		return 0, nil
	}
	contentLength, err := strconv.Atoi(contentLengthString)
	if err != nil {
		return 0, err
	}
	if len(data)+len(r.Body) > contentLength {
		return 0, fmt.Errorf("content length is invalid")
	}
	r.Body = append(r.Body, data...)
	if len(r.Body) == contentLength {
		r.State = DoneState
	}
	return len(data), nil
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

func parseRequestFromReader(reader io.Reader, buf []byte, dataBuffer []byte, r *Request) error {
	for r.State != DoneState {
		n, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n > 0 {
			dataBuffer = append(dataBuffer, buf[:n]...)
		}
		consumed, parseErr := r.parse(dataBuffer)
		if parseErr != nil {
			return parseErr
		}
		if consumed > 0 {
			dataBuffer = dataBuffer[consumed:]
		}
		if err == io.EOF {
			break
		}
	}
	if r.State != DoneState {
		return ErrBadRequest
	}
	return nil
}
