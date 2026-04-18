package research

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShellQuote(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain_username",
			input:    "username",
			expected: "'username'",
		},
		{
			name:     "simple_name",
			input:    "alice",
			expected: "'alice'",
		},
		{
			name:     "empty_string",
			input:    "",
			expected: "''",
		},
		{
			name:     "embedded_single_quote",
			input:    "bob's",
			expected: "'bob'\\''s'",
		},
		{
			name:     "multiple_embedded_single_quotes",
			input:    "a'b'c",
			expected: "'a'\\''b'\\''c'",
		},
		{
			name:     "semicolon_injection_attempt",
			input:    "user; rm -rf /",
			expected: "'user; rm -rf /'",
		},
		{
			name:     "command_substitution_attempt",
			input:    "$(whoami)",
			expected: "'$(whoami)'",
		},
		{
			name:     "backtick_injection_attempt",
			input:    "`id`",
			expected: "'`id`'",
		},
		{
			name:     "pipe_injection_attempt",
			input:    "user|cat /etc/passwd",
			expected: "'user|cat /etc/passwd'",
		},
		{
			name:     "ampersand_injection_attempt",
			input:    "user&background",
			expected: "'user&background'",
		},
		{
			name:     "newline_injection_attempt",
			input:    "user\nmalicious",
			expected: "'user\nmalicious'",
		},
		{
			name:     "path_with_spaces",
			input:    "/home/my user/file",
			expected: "'/home/my user/file'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shellQuote(tt.input)
			assert.Equal(t, tt.expected, result,
				"shellQuote(%q) = %q, want %q", tt.input, result, tt.expected)
		})
	}
}

// TestShellQuoteStructuralInvariants verifies that shellQuote output always:
//  1. Starts with a single quote
//  2. Ends with a single quote
//  3. Escapes embedded single quotes using the POSIX '\” sequence
//     (each ' in input adds exactly one '\” in output)
func TestShellQuoteStructuralInvariants(t *testing.T) {
	inputs := []string{
		"'; rm -rf /; echo '",
		"' OR '1'='1",
		"`whoami`",
		"$(cat /etc/shadow)",
		"a\nb",
		"normal",
		"",
	}

	for _, input := range inputs {
		quoted := shellQuote(input)

		assert.True(t, len(quoted) >= 2,
			"output must be at least 2 chars for input %q", input)
		assert.Equal(t, "'", string(quoted[0]),
			"output must start with single quote for input %q", input)
		assert.Equal(t, "'", string(quoted[len(quoted)-1]),
			"output must end with single quote for input %q", input)

		// Count single quotes in input; each should produce one '\'' in output.
		inputQuotes := 0
		for _, c := range input {
			if c == '\'' {
				inputQuotes++
			}
		}
		outputEscapes := strings.Count(quoted, "'\\''")
		assert.Equal(t, inputQuotes, outputEscapes,
			"each input single quote must produce one '\\'' escape for input %q", input)
	}
}
