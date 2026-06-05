package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

const maxTagNameLength = 1024

func validateADIF(data []byte) error {
	_, err := validateADIFWithWarnings(data)
	return err
}

func validateADIFWithWarnings(data []byte) ([]string, error) {
	inHeader := hasADIFHeader(data)
	var warnings []string

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
			return nil, errorAtf(data, i, "stray data byte %q", b)
		}

		end := i + 1
		for end < len(data) && data[end] != '>' {
			end++
		}
		if end >= len(data) {
			return nil, errorAtf(data, i, "unterminated tag")
		}

		tagText := string(data[i+1 : end])
		length, dataType, err := parseTag(tagText)
		if err != nil {
			return nil, errorAtf(data, i, "invalid tag: %w", err)
		}

		i = end + 1
		if length > 0 {
			if i+length > len(data) {
				return nil, errorAtf(data, i, "tag data exceeds input")
			}
			if strings.EqualFold(dataType, "M") {
				warnings = append(warnings, multilineLineEndingWarnings(data, i, data[i:i+length])...)
			}
			i += length
		}

		if inHeader && strings.EqualFold(tagText, "EOH") {
			inHeader = false
		}
	}

	return warnings, nil
}

func errorAtf(data []byte, offset int, format string, args ...any) error {
	line, column := lineColumn(data, offset)
	positionedFormat := fmt.Sprintf("%s at line %d, column %d", format, line, column)
	return fmt.Errorf(positionedFormat, args...)
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
			if i+1 < len(data) && data[i+1] == '\n' {
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
	length, _, err := parseTag(tag)
	return length, err
}

func parseTag(tag string) (int, string, error) {
	parts := strings.Split(tag, ":")
	if parts[0] == "" {
		return 0, "", fmt.Errorf("empty tag name")
	}
	if len(parts) > 3 {
		return 0, "", fmt.Errorf("too many tag segments")
	}
	if len(parts[0]) > maxTagNameLength {
		return 0, "", fmt.Errorf("tag name too long")
	}

	for _, r := range parts[0] {
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_') {
			return 0, "", fmt.Errorf("invalid tag name %q", parts[0])
		}
	}

	if len(parts) == 1 {
		return 0, "", nil
	}
	if parts[1] == "" {
		return 0, "", fmt.Errorf("missing run length")
	}

	length, err := strconv.Atoi(parts[1])
	if err != nil || length < 0 {
		return 0, "", fmt.Errorf("invalid run length %q", parts[1])
	}
	dataType := ""
	if len(parts) == 3 && parts[2] == "" {
		return 0, "", fmt.Errorf("empty data type")
	}
	if len(parts) == 3 {
		dataType = parts[2]
	}

	return length, dataType, nil
}

func isWhitespaceByte(b byte) bool {
	return b == ' ' || b == '\n' || b == '\r' || b == '\t' || b == '\v' || b == '\f'
}

func multilineLineEndingWarnings(data []byte, fieldStart int, fieldData []byte) []string {
	var warnings []string
	for i := 0; i < len(fieldData); i++ {
		switch fieldData[i] {
		case '\r':
			if i+1 < len(fieldData) && fieldData[i+1] == '\n' {
				i++
				continue
			}
			warnings = append(warnings, warningAtf(data, fieldStart+i, "non-CRLF line ending in multiline field"))
		case '\n':
			warnings = append(warnings, warningAtf(data, fieldStart+i, "non-CRLF line ending in multiline field"))
		}
	}

	return warnings
}

func warningAtf(data []byte, offset int, message string) string {
	line, column := lineColumn(data, offset)
	return fmt.Sprintf("%s at line %d, column %d", message, line, column)
}
