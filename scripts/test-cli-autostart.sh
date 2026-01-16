#!/bin/bash
# Test CLI auto-start functionality

set -e

# Cleanup - ensure no zombie processes (use -9 for force kill)
pkill -9 -f "bin/prismd" 2>/dev/null || true
# Wait longer to ensure port is released
sleep 3

echo "Testing CLI auto-start..."

# Run CLI command - should auto-start daemon
timeout 15s ./bin/prism workspace list > /dev/null 2>&1

echo "✅ CLI auto-start working"

# Cleanup
pkill -9 -f "bin/prismd" 2>/dev/null || true
