package main

import (
	"strings"
	"testing"
)

func TestValidateADIF(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantErr         bool
		wantErrContains string
	}{
		{
			name:    "valid field and eor",
			input:   "<CALL:5>ABCDE<EOR>",
			wantErr: false,
		},
		{
			name:    "valid with type and whitespace",
			input:   " \n<CALL:3:S>ABC\t<EOR>\n",
			wantErr: false,
		},
		{
			name:    "valid with additional ascii whitespace",
			input:   "\v<CALL:3>ABC\f<EOR>\v",
			wantErr: false,
		},
		{
			name: "valid header with free text and header tags",
			input: "QRZLogbook download for k0swe\n" +
				"    Date: Fri Jun  5 22:39:53 2026\n" +
				"    Bookid: 230915\n" +
				"    Records: 1417\n" +
				"    <ADIF_VER:5>3.1.1\n" +
				"    <PROGRAMID:10>QRZLogbook\n" +
				"    <PROGRAMVERSION:3>2.0\n" +
				"    <eoh>\n" +
				"<CALL:5>K0SWE<EOR>",
			wantErr: false,
		},
		{
			name:            "stray byte after data",
			input:           "<CALL:3>ABCx<EOR>",
			wantErr:         true,
			wantErrContains: "line 1, column 12",
		},
		{
			name:            "stray text after eoh is invalid",
			input:           "Header text<EOH>\noops<CALL:5>K0SWE<EOR>",
			wantErr:         true,
			wantErrContains: "line 2, column 1",
		},
		{
			name:            "unterminated tag",
			input:           "<CALL:3>ABC<EOR",
			wantErr:         true,
			wantErrContains: "line 1, column 12",
		},
		{
			name:            "invalid run length",
			input:           "\r\n<CALL:xx>ABC",
			wantErr:         true,
			wantErrContains: "line 2, column 1",
		},
		{
			name:    "missing run length",
			input:   "<CALL:>ABC",
			wantErr: true,
		},
		{
			name:    "empty data type",
			input:   "<CALL:3:>ABC",
			wantErr: true,
		},
		{
			name:    "too many tag segments",
			input:   "<CALL:3:S:EXTRA>ABC",
			wantErr: true,
		},
		{
			name:            "truncated data",
			input:           "<CALL:5>ABC",
			wantErr:         true,
			wantErrContains: "line 1, column 9",
		},
		{
			name:    "empty tag name",
			input:   "<:3>ABC",
			wantErr: true,
		},
		{
			name:    "tag name at max length",
			input:   "<" + strings.Repeat("A", maxTagNameLength) + ":1>A",
			wantErr: false,
		},
		{
			name:    "tag name too long",
			input:   "<" + strings.Repeat("A", maxTagNameLength+1) + ":1>A",
			wantErr: true,
		},
	}

	func TestLineColumn(t *testing.T) {
		data := []byte("A\r\nB\nC")
		tests := []struct {
			offset     int
			wantLine   int
			wantColumn int
		}{
			{offset: 0, wantLine: 1, wantColumn: 1},
			{offset: 1, wantLine: 1, wantColumn: 2},
			{offset: 2, wantLine: 2, wantColumn: 1},
			{offset: 3, wantLine: 2, wantColumn: 1},
			{offset: 5, wantLine: 3, wantColumn: 1},
		}

		for _, tc := range tests {
			line, column := lineColumn(data, tc.offset)
			if line != tc.wantLine || column != tc.wantColumn {
				t.Fatalf("lineColumn(%d) = (%d, %d), want (%d, %d)", tc.offset, line, column, tc.wantLine, tc.wantColumn)
			}
		}
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateADIF([]byte(tc.input))
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateADIF() error = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErrContains != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErrContains)) {
				t.Fatalf("validateADIF() error = %v, want substring %q", err, tc.wantErrContains)
			}
		})
	}
}
