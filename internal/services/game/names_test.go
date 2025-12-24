package game

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanonicalPlayerName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase name",
			input:    "testuser",
			expected: "testuser",
		},
		{
			name:     "uppercase name",
			input:    "TESTUSER",
			expected: "testuser",
		},
		{
			name:     "mixed case name",
			input:    "TestUser",
			expected: "testuser",
		},
		{
			name:     "name with leading spaces",
			input:    "  testuser",
			expected: "testuser",
		},
		{
			name:     "name with trailing spaces",
			input:    "testuser  ",
			expected: "testuser",
		},
		{
			name:     "name with leading and trailing spaces",
			input:    "  testuser  ",
			expected: "testuser",
		},
		{
			name:     "empty name",
			input:    "",
			expected: "",
		},
		{
			name:     "only spaces",
			input:    "   ",
			expected: "",
		},
		{
			name:     "name with numbers",
			input:    "User123",
			expected: "user123",
		},
		{
			name:     "name with special characters",
			input:    "User_Name-123",
			expected: "user_name-123",
		},
		{
			name:     "name with unicode",
			input:    "UserÄÖÜ",
			expected: "useräöü", // ToLower handles unicode
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := canonicalPlayerName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
