package headers

import (
	"fmt"
	"strings"
)

const (
	colonNotFound       = -1
	firstCharacterIndex = 0
	spaceCharacter      = ' '
)

func hasSpaceBeforeColon(s string) bool {
	idx := strings.IndexRune(s, colonRune)
	if idx == colonNotFound || idx == firstCharacterIndex {
		return false
	}
	if s[idx-1] == spaceCharacter {
		return true
	}
	return false
}

func validateHeaderKey(rawKey string) error {
	for _, r := range rawKey {
		if r <= asciiSpace || r > asciiTilde || r == colonRune {
			return fmt.Errorf("invalid character in header key: %v", rawKey)
		}
	}
	return nil
}
