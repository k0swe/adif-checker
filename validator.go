package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const maxTagNameLength = 1024

func validateADIF(data []byte) error {
	for i := 0; i < len(data); {
		b := data[i]
		if isWhitespaceByte(b) {
			i++
			continue
		}

		if b != '<' {
			return fmt.Errorf("stray data byte %q at offset %d", b, i)
		}

		end := i + 1
		for end < len(data) && data[end] != '>' {
			end++
		}
		if end >= len(data) {
			return fmt.Errorf("unterminated tag at offset %d", i)
		}

		tagText := string(data[i+1 : end])
		length, err := parseTagLength(tagText)
		if err != nil {
			return fmt.Errorf("invalid tag at offset %d: %w", i, err)
		}

		i = end + 1
		if length > 0 {
			if i+length > len(data) {
				return fmt.Errorf("tag data exceeds input at offset %d", i)
			}
			i += length
		}
	}

	return nil
}

func parseTagLength(tag string) (int, error) {
	parts := strings.Split(tag, ":")
	if parts[0] == "" {
		return 0, fmt.Errorf("empty tag name")
	}
	if len(parts) > 3 {
		return 0, fmt.Errorf("too many tag segments")
	}
	if len(parts[0]) > maxTagNameLength {
		return 0, fmt.Errorf("tag name too long")
	}

	for _, r := range parts[0] {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_') {
			return 0, fmt.Errorf("invalid tag name %q", parts[0])
		}
	}

	if len(parts) == 1 {
		return 0, nil
	}
	if parts[1] == "" {
		return 0, fmt.Errorf("missing run length")
	}

	length, err := strconv.Atoi(parts[1])
	if err != nil || length < 0 {
		return 0, fmt.Errorf("invalid run length %q", parts[1])
	}
	if len(parts) == 3 && parts[2] == "" {
		return 0, fmt.Errorf("empty data type")
	}

	return length, nil
}

func isWhitespaceByte(b byte) bool {
	return b == ' ' || b == '\n' || b == '\r' || b == '\t' || b == '\v' || b == '\f'
}
