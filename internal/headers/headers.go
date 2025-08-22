package headers

import (
	"fmt"
	"strings"
)

type Headers map[string]string

func NewHeaders() Headers {
	return make(Headers)
}

func hasSpaceBeforeColon(s string) bool {
	idx := strings.IndexRune(s, ':')
	if idx == -1 || idx == 0 {
		return false
	}
	if s[idx-1] == ' ' {
		return true
	}
	return false
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	headerString := string(data)

	if strings.HasPrefix(headerString, "\r\n\r\n") {
		return len("\r\n\r\n"), true, nil
	}

	endIdx := strings.Index(headerString, "\r\n\r\n")
	if endIdx == -1 {
		return 0, false, nil
	}

	headerBlock := headerString[:endIdx]
	headerLines := strings.Split(headerBlock, "\r\n")

	for _, headerLine := range headerLines {
		if headerLine == "" {
			continue
		}
		if hasSpaceBeforeColon(headerLine) {
			return 0, false, fmt.Errorf("has a space before colon, header: %v", headerLine)
		}
		kv := strings.SplitN(headerLine, ":", 2)
		if len(kv) != 2 {
			return 0, false, fmt.Errorf("malinformed header: %v lenkv: %v", headerLine, len(kv))
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		h[key] = value
	}
	return endIdx + len("\r\n\r\n"), true, nil
}
