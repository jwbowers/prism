package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogLevels(t *testing.T) {
	tests := []struct {
		name           string
		level          LogLevel
		logFunc        func(string, ...any)
		message        string
		expectStdout   bool
		expectStderr   bool
		expectedOutput string
	}{
		{
			name:           "Debug message",
			level:          LevelDebug,
			logFunc:        Debug,
			message:        "debug test",
			expectStdout:   true,
			expectStderr:   false,
			expectedOutput: "level=DEBUG msg=\"debug test\"",
		},
		{
			name:           "Info message",
			level:          LevelInfo,
			logFunc:        Info,
			message:        "info test",
			expectStdout:   true,
			expectStderr:   false,
			expectedOutput: "level=INFO msg=\"info test\"",
		},
		{
			name:           "Warn message",
			level:          LevelInfo,
			logFunc:        Warn,
			message:        "warn test",
			expectStdout:   false,
			expectStderr:   true,
			expectedOutput: "level=WARN msg=\"warn test\"",
		},
		{
			name:           "Error message",
			level:          LevelInfo,
			logFunc:        Error,
			message:        "error test",
			expectStdout:   false,
			expectStderr:   true,
			expectedOutput: "level=ERROR msg=\"error test\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create buffers to capture output
			stdoutBuf := &bytes.Buffer{}
			stderrBuf := &bytes.Buffer{}

			// Initialize with test level
			Initialize(tt.level)
			SetOutput(stdoutBuf, stderrBuf)

			// Call the log function
			tt.logFunc(tt.message)

			// Check stdout
			stdoutOutput := stdoutBuf.String()
			if tt.expectStdout {
				if !strings.Contains(stdoutOutput, tt.expectedOutput) {
					t.Errorf("Expected stdout to contain %q, got %q", tt.expectedOutput, stdoutOutput)
				}
			} else {
				if stdoutOutput != "" {
					t.Errorf("Expected no stdout output, got %q", stdoutOutput)
				}
			}

			// Check stderr
			stderrOutput := stderrBuf.String()
			if tt.expectStderr {
				if !strings.Contains(stderrOutput, tt.expectedOutput) {
					t.Errorf("Expected stderr to contain %q, got %q", tt.expectedOutput, stderrOutput)
				}
			} else {
				if stderrOutput != "" {
					t.Errorf("Expected no stderr output, got %q", stderrOutput)
				}
			}
		})
	}
}

func TestLogWithAttributes(t *testing.T) {
	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}

	Initialize(LevelDebug)
	SetOutput(stdoutBuf, stderrBuf)

	// Test logging with key-value attributes
	Info("operation complete", "status", "success", "duration_ms", 150)

	output := stdoutBuf.String()
	if !strings.Contains(output, "operation complete") {
		t.Errorf("Expected message in output")
	}
	if !strings.Contains(output, "status=success") {
		t.Errorf("Expected status attribute in output")
	}
	if !strings.Contains(output, "duration_ms=150") {
		t.Errorf("Expected duration_ms attribute in output")
	}
}

func TestGetLevel(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		expected LogLevel
	}{
		{"Debug level", "debug", LevelDebug},
		{"Debug level uppercase", "DEBUG", LevelDebug},
		{"Info level", "info", LevelInfo},
		{"Info level uppercase", "INFO", LevelInfo},
		{"Warn level", "warn", LevelWarn},
		{"Warn level uppercase", "WARN", LevelWarn},
		{"Warning level", "warning", LevelWarn},
		{"Error level", "error", LevelError},
		{"Error level uppercase", "ERROR", LevelError},
		{"Default level", "", LevelInfo},
		{"Invalid level", "invalid", LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envValue != "" {
				t.Setenv("PRISM_LOG_LEVEL", tt.envValue)
			}

			level := GetLevel()
			if level != tt.expected {
				t.Errorf("Expected level %q, got %q", tt.expected, level)
			}
		})
	}
}

func TestDebugFilteredAtInfoLevel(t *testing.T) {
	stdoutBuf := &bytes.Buffer{}
	stderrBuf := &bytes.Buffer{}

	// Reset and initialize at INFO level
	Reset()
	Initialize(LevelInfo)
	SetOutput(stdoutBuf, stderrBuf)

	// Log a debug message
	Debug("this should not appear")

	output := stdoutBuf.String()
	if output != "" {
		t.Errorf("Expected no output for DEBUG message at INFO level, got %q", output)
	}
}
