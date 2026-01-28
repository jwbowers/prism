// Package logger provides structured logging for the Prism daemon with proper log levels.
//
// Log levels and their destinations:
//   - DEBUG: Detailed diagnostic information → stdout
//   - INFO: Normal operation messages → stdout
//   - WARN: Warning conditions → stderr
//   - ERROR: Error conditions → stderr
//
// Usage:
//
//	logger.Info("Server starting", "port", 8947)
//	logger.Debug("Processing request", "method", "GET", "path", "/api/v1/instances")
//	logger.Warn("Rate limit approaching", "current", 95, "max", 100)
//	logger.Error("Failed to connect", "error", err)
package logger

import (
	"io"
	"log/slog"
	"os"
	"sync"
)

var (
	// Global logger instances
	stdoutLogger *slog.Logger
	stderrLogger *slog.Logger
	once         sync.Once
	currentLevel slog.Level = slog.LevelInfo
)

// LogLevel represents the logging level
type LogLevel string

const (
	LevelDebug LogLevel = "debug"
	LevelInfo  LogLevel = "info"
	LevelWarn  LogLevel = "warn"
	LevelError LogLevel = "error"
)

// Initialize sets up the logging infrastructure
// This should be called once at daemon startup
func Initialize(level LogLevel) {
	once.Do(func() {
		setupLoggers(level)
	})
}

// Reset resets the logger state (primarily for testing)
func Reset() {
	once = sync.Once{}
	stdoutLogger = nil
	stderrLogger = nil
	currentLevel = slog.LevelInfo
}

// setupLoggers creates stdout and stderr loggers with appropriate levels
func setupLoggers(level LogLevel) {
	// Map string level to slog.Level
	var slogLevel slog.Level
	switch level {
	case LevelDebug:
		slogLevel = slog.LevelDebug
	case LevelInfo:
		slogLevel = slog.LevelInfo
	case LevelWarn:
		slogLevel = slog.LevelWarn
	case LevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	// Store current level
	currentLevel = slogLevel

	// Create handlers with custom formatting
	stdoutHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slogLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			// Simplify time format
			if a.Key == slog.TimeKey {
				return slog.Attr{Key: "time", Value: slog.StringValue(a.Value.Time().Format("2006/01/02 15:04:05"))}
			}
			return a
		},
	})

	stderrHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelWarn, // Only WARN and ERROR go to stderr
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{Key: "time", Value: slog.StringValue(a.Value.Time().Format("2006/01/02 15:04:05"))}
			}
			return a
		},
	})

	stdoutLogger = slog.New(stdoutHandler)
	stderrLogger = slog.New(stderrHandler)
}

// ensureInitialized ensures loggers are initialized with default settings
func ensureInitialized() {
	if stdoutLogger == nil || stderrLogger == nil {
		Initialize(LevelInfo)
	}
}

// Debug logs a debug message to stdout
func Debug(msg string, args ...any) {
	ensureInitialized()
	stdoutLogger.Debug(msg, args...)
}

// Info logs an informational message to stdout
func Info(msg string, args ...any) {
	ensureInitialized()
	stdoutLogger.Info(msg, args...)
}

// Warn logs a warning message to stderr
func Warn(msg string, args ...any) {
	ensureInitialized()
	stderrLogger.Warn(msg, args...)
}

// Error logs an error message to stderr
func Error(msg string, args ...any) {
	ensureInitialized()
	stderrLogger.Error(msg, args...)
}

// Fatal logs an error message and exits with status 1
func Fatal(msg string, args ...any) {
	ensureInitialized()
	stderrLogger.Error(msg, args...)
	os.Exit(1)
}

// SetOutput redirects logger output (primarily for testing)
// Preserves the current log level settings
func SetOutput(stdout, stderr io.Writer) {
	stdoutHandler := slog.NewTextHandler(stdout, &slog.HandlerOptions{
		Level: currentLevel, // Preserve current level
	})
	stderrHandler := slog.NewTextHandler(stderr, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	})

	stdoutLogger = slog.New(stdoutHandler)
	stderrLogger = slog.New(stderrHandler)
}

// GetLevel returns the current log level from environment or default
func GetLevel() LogLevel {
	level := os.Getenv("PRISM_LOG_LEVEL")
	switch level {
	case "debug", "DEBUG":
		return LevelDebug
	case "info", "INFO":
		return LevelInfo
	case "warn", "WARN", "warning", "WARNING":
		return LevelWarn
	case "error", "ERROR":
		return LevelError
	default:
		return LevelInfo
	}
}
