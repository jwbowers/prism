#!/bin/bash
# fix-dependencies.sh - Repairs and updates Go module dependencies
# Addresses issues found by security scanning tools

set -e

echo "🔍 Checking dependency structure..."

# Check if go.sum is out of sync with go.mod
if ! go mod verify &>/dev/null; then
  echo "⚠️  go.sum is out of sync with go.mod"
  
  # Make backup of current dependency files
  echo "📦 Backing up current dependency files..."
  cp go.mod go.mod.backup
  if [ -f go.sum ]; then
    cp go.sum go.sum.backup
  fi
  
  # Download missing dependencies
  echo "📥 Downloading missing dependencies..."
  go mod download
  
  # Clean up and regenerate dependency information
  echo "🧹 Tidying dependencies..."
  go mod tidy
fi

# Check for vulnerable dependencies with govulncheck
echo "🔒 Scanning for vulnerable dependencies..."
if command -v govulncheck &>/dev/null; then
  if ! govulncheck -show=package ./... 2>/dev/null; then
    echo "⚠️  Potential vulnerabilities detected"
    echo "📋 For detailed vulnerability information, run: govulncheck -v ./..."
  else
    echo "✅ No vulnerabilities detected"
  fi
else
  echo "⚠️  govulncheck not found, skipping vulnerability scan"
  echo "💡 Install govulncheck with: go install golang.org/x/vuln/cmd/govulncheck@latest"
fi

# Run tests to ensure dependencies work correctly
echo "🧪 Verifying build with dependencies..."
if go build -o /dev/null ./cmd/prism 2>/dev/null; then
  echo "✅ CLI client builds successfully"
else
  echo "❌ CLI client build failed"
fi

if go build -o /dev/null ./cmd/prismd 2>/dev/null; then
  echo "✅ Daemon builds successfully"
else
  echo "❌ Daemon build failed"
fi

echo "✨ Dependency check complete"