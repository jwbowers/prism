#!/bin/bash
# Clean up all test-* instances from current session
# Usage: ./scripts/cleanup-current-tests.sh

set -e

PRISM="./bin/prism"

if [ ! -f "$PRISM" ]; then
    echo "❌ Error: prism binary not found at $PRISM"
    echo "   Run 'make build' first"
    exit 1
fi

echo "🧹 Cleaning up all test-* instances..."
echo ""

# Get list of test instances
TEST_INSTANCES=$($PRISM workspace list | grep "^test-" | awk '{print $1}' || true)

if [ -z "$TEST_INSTANCES" ]; then
    echo "✅ No test instances found"
    exit 0
fi

# Count instances
COUNT=$(echo "$TEST_INSTANCES" | wc -l | tr -d ' ')
echo "Found $COUNT test instance(s) to delete:"
echo "$TEST_INSTANCES"
echo ""

# Confirm deletion
read -p "Delete all test instances? (y/n) " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Cancelled"
    exit 1
fi

# Delete each instance
SUCCESS=0
FAILED=0

echo "$TEST_INSTANCES" | while read -r name; do
    echo -n "Deleting: $name... "
    if $PRISM workspace delete "$name" >/dev/null 2>&1; then
        echo "✅"
        ((SUCCESS++))
    else
        echo "❌"
        ((FAILED++))
    fi
done

echo ""
echo "📊 Cleanup complete"
echo "   Run 'prism workspace list' to verify"
