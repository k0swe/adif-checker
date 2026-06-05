package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const maxTagNameLength = 1024

func validateADIF(data []byte) error {
	inHeader := hasADIFHeader(data)

	for i := 0; i < len(data); {
		b := data[i]
		if isWhitespaceByte(b) {
			i++
			continue
		}

		if b != '<' {
			if inHeader {
				i++
				continue
			}
			return errorAtf(data, i, "stray data byte %q", b)
		}

		end := i + 1
		for end < len(data) && data[end] != '>' {
			end++
		}
		if end >= len(data) {
			return errorAtf(data, i, "unterminated tag")
		}

		tagText := string(data[i+1 : end])
		length, err := parseTagLength(tagText)
		if err != nil {
			return errorAtf(data, i, "invalid tag: %w", err)
		}

		i = end + 1
		if length > 0 {
			if i+length > len(data) {
				return errorAtf(data, i, "tag data exceeds input")
			}
			i += length
		}

		if inHeader && strings.EqualFold(tagText, "EOH") {
			inHeader = false
		}
	}

	return nil
}

func errorAtf(data []byte, offset int, format string, args ...any) error {
	line, column := lineColumn(data, offset)
	allArgs := append(append([]any{}, args...), line, column)
	return fmt.Errorf(format+" at line %d, column %d", allArgs...)
}

func lineColumn(data []byte, offset int) (line int, column int) {
	line = 1
	column = 1
	if offset <= 0 {
		return line, column
	}
	if offset > len(data) {
		offset = len(data)
	}

	for i := 0; i < offset; i++ {
		switch data[i] {
		case '\r':
			line++
			column = 1
			if i+1 < offset && data[i+1] == '\n' {
				i++
			}
		case '\n':
			line++
			column = 1
		default:
			column++
		}
	}

	return line, column
}

func hasADIFHeader(data []byte) bool {
	for _, b := range data {
		if isWhitespaceByte(b) {
			continue
		}
		return b != '<'
	}
	return false
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
