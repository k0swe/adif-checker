package main

import (
	"strings"
	"testing"
)

func TestValidateADIF(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
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
			name:    "stray byte after data",
			input:   "<CALL:3>ABCx<EOR>",
			wantErr: true,
		},
		{
			name:    "unterminated tag",
			input:   "<CALL:3>ABC<EOR",
			wantErr: true,
		},
		{
			name:    "invalid run length",
			input:   "<CALL:xx>ABC",
			wantErr: true,
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
			name:    "truncated data",
			input:   "<CALL:5>ABC",
			wantErr: true,
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

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateADIF([]byte(tc.input))
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateADIF() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}
