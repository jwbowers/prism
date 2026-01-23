//go:build integration
// +build integration

package integration

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// CLITestContext provides a test context for CLI command execution
type CLITestContext struct {
	t           *testing.T
	tempDir     string
	prismBinary string
	env         []string
}

// NewCLITestContext creates a new CLI test context
func NewCLITestContext(t *testing.T) *CLITestContext {
	tempDir := t.TempDir()

	// Find prism binary
	prismBinary, err := findPrismBinary()
	require.NoError(t, err, "Failed to find prism binary")

	// Set up environment with isolated config
	env := os.Environ()
	env = append(env,
		fmt.Sprintf("HOME=%s", tempDir),
		fmt.Sprintf("PRISM_CONFIG_DIR=%s/.prism", tempDir),
	)

	ctx := &CLITestContext{
		t:           t,
		tempDir:     tempDir,
		prismBinary: prismBinary,
		env:         env,
	}

	// Ensure config directory exists
	configDir := filepath.Join(tempDir, ".prism")
	err = os.MkdirAll(configDir, 0755)
	require.NoError(t, err, "Failed to create config directory")

	return ctx
}

// RunCommand executes a CLI command and returns output
func (ctx *CLITestContext) RunCommand(args ...string) (string, error) {
	cmd := exec.Command(ctx.prismBinary, args...)
	cmd.Env = ctx.env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()
	if err != nil {
		// Include stderr in error for debugging
		if stderr.Len() > 0 {
			output = fmt.Sprintf("STDOUT:\n%s\n\nSTDERR:\n%s", stdout.String(), stderr.String())
		}
		return output, fmt.Errorf("command failed: %w", err)
	}

	return output, nil
}

// RunCommandWithInput executes a CLI command with stdin input
func (ctx *CLITestContext) RunCommandWithInput(input string, args ...string) (string, error) {
	cmd := exec.Command(ctx.prismBinary, args...)
	cmd.Env = ctx.env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Stdin = strings.NewReader(input)

	err := cmd.Run()
	output := stdout.String()
	if err != nil {
		// Include stderr in error for debugging
		if stderr.Len() > 0 {
			output = fmt.Sprintf("STDOUT:\n%s\n\nSTDERR:\n%s", stdout.String(), stderr.String())
		}
		return output, fmt.Errorf("command failed: %w", err)
	}

	return output, nil
}

// RunCommandExpectError executes a CLI command expecting it to fail
func (ctx *CLITestContext) RunCommandExpectError(args ...string) (string, error) {
	cmd := exec.Command(ctx.prismBinary, args...)
	cmd.Env = ctx.env

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := stdout.String()
	if stderr.Len() > 0 {
		output = fmt.Sprintf("%s\n%s", output, stderr.String())
	}

	return output, err
}

// GetConfigDir returns the isolated config directory
func (ctx *CLITestContext) GetConfigDir() string {
	return filepath.Join(ctx.tempDir, ".prism")
}

// GetCommunityDir returns the community templates directory
func (ctx *CLITestContext) GetCommunityDir() string {
	return filepath.Join(ctx.GetConfigDir(), "community")
}

// WriteFile writes content to a file in the temp directory
func (ctx *CLITestContext) WriteFile(relativePath string, content string) error {
	fullPath := filepath.Join(ctx.tempDir, relativePath)
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(fullPath, []byte(content), 0644)
}

// ReadFile reads a file from the temp directory
func (ctx *CLITestContext) ReadFile(relativePath string) (string, error) {
	fullPath := filepath.Join(ctx.tempDir, relativePath)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// FileExists checks if a file exists in the temp directory
func (ctx *CLITestContext) FileExists(relativePath string) bool {
	fullPath := filepath.Join(ctx.tempDir, relativePath)
	_, err := os.Stat(fullPath)
	return err == nil
}

// findPrismBinary locates the prism binary
func findPrismBinary() (string, error) {
	// Try common locations
	locations := []string{
		"./bin/prism",
		"../bin/prism",
		"../../bin/prism",
		"../../../bin/prism",
	}

	// Also check if prism is in PATH
	if path, err := exec.LookPath("prism"); err == nil {
		return path, nil
	}

	// Check relative locations
	for _, loc := range locations {
		absPath, err := filepath.Abs(loc)
		if err != nil {
			continue
		}
		if _, err := os.Stat(absPath); err == nil {
			return absPath, nil
		}
	}

	return "", fmt.Errorf("prism binary not found. Build it with 'make build' or 'go build -o bin/prism ./cmd/prism'")
}

// CaptureOutput captures stdout and returns it
func CaptureOutput(f func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}
