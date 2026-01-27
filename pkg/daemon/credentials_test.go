// Package daemon provides credential error detection tests
package daemon

import (
	"errors"
	"strings"
	"testing"
)

// TestIsCredentialError tests credential error detection
func TestIsCredentialError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "no EC2 IMDS role",
			err:      errors.New("no ec2 imds role found"),
			expected: true,
		},
		{
			name:     "failed to refresh credentials",
			err:      errors.New("failed to refresh cached credentials"),
			expected: true,
		},
		{
			name:     "no valid providers",
			err:      errors.New("no valid providers in chain"),
			expected: true,
		},
		{
			name:     "credentials not found",
			err:      errors.New("credentials not found"),
			expected: true,
		},
		{
			name:     "IMDS timeout",
			err:      errors.New("operation error ec2imds: Get metadata request failed"),
			expected: true,
		},
		{
			name:     "EC2 IMDS dial timeout",
			err:      errors.New("dial tcp 169.254.169.254:80: i/o timeout"),
			expected: true,
		},
		{
			name:     "generic i/o timeout",
			err:      errors.New("i/o timeout"),
			expected: true,
		},
		{
			name:     "unrelated error",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "network error",
			err:      errors.New("connection refused"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isCredentialError(tt.err)
			if result != tt.expected {
				t.Errorf("isCredentialError(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}

// TestGetCredentialErrorMessage tests error message generation
func TestGetCredentialErrorMessage(t *testing.T) {
	msg := getCredentialErrorMessage()
	if msg == "" {
		t.Error("getCredentialErrorMessage() returned empty string")
	}

	// Check that message contains key information
	requiredPhrases := []string{
		"AWS credentials required",
		"aws configure",
		"AWS_ACCESS_KEY_ID",
		"prism profile create",
	}

	for _, phrase := range requiredPhrases {
		if !strings.Contains(msg, phrase) {
			t.Errorf("getCredentialErrorMessage() missing required phrase: %s", phrase)
		}
	}
}

// TestGetReducedModeBanner tests banner message generation
func TestGetReducedModeBanner(t *testing.T) {
	banner := getReducedModeBanner()
	if banner == "" {
		t.Error("getReducedModeBanner() returned empty string")
	}

	// Check that banner contains key information
	requiredPhrases := []string{
		"reduced functionality mode",
		"Template validation",
		"Instance operations",
		"aws configure",
	}

	for _, phrase := range requiredPhrases {
		if !strings.Contains(banner, phrase) {
			t.Errorf("getReducedModeBanner() missing required phrase: %s", phrase)
		}
	}
}
