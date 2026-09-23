package repository

import (
	"testing"
)

func TestFormatTimeHi(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"0800", "08:00"},
		{"0900", "09:00"},
		{"08:00", "08:00"},
		{"09:30", "09:30"},
		{"08:00:00", "08:00"},
		{"800", "08:00"},
		{"8:00", "08:00"},
		{"1345", "13:45"},
		{"", ""},
		{"   ", ""},
	}

	for _, tt := range tests {
		result := formatTimeHi(tt.input)
		if result != tt.expected {
			t.Errorf("formatTimeHi(%q) = %q; want %q", tt.input, result, tt.expected)
		}
	}
}
