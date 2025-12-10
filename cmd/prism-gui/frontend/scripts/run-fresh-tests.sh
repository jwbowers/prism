#!/bin/bash
set -e

echo "🧹 Clearing test caches and ensuring fresh code..."
echo ""

# Get the script directory and navigate to frontend root
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
FRONTEND_DIR="$( cd "$SCRIPT_DIR/.." && pwd )"

cd "$FRONTEND_DIR"

echo "Working directory: $FRONTEND_DIR"
echo ""

echo "1. Clearing Playwright cache..."
rm -rf node_modules/.cache

echo "2. Clearing build artifacts..."
rm -rf dist/ .next/ build/ playwright-report/

echo "3. Killing any running processes..."
pkill -f playwright || true
pkill -f prismd || true
pkill -f prism-gui || true

echo "4. Verifying git status (checking for uncommitted changes)..."
git status --short

echo ""
echo "5. Running tests with fresh code..."
echo "   Target: ${1:-tests/e2e/invitation-workflows.spec.ts}"
echo ""

# Default to invitation-workflows if no argument provided
TEST_TARGET="${1:-tests/e2e/invitation-workflows.spec.ts}"

npx playwright test "$TEST_TARGET" --project=chromium --reporter=list

echo ""
echo "✅ Test run complete!"
