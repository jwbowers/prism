#!/bin/bash
# Alert if test instance costs exceed threshold
# Usage: ./scripts/check-test-costs.sh [threshold]

set -e

PRISM="./bin/prism"
THRESHOLD=${1:-50}  # Default $50/day threshold

if [ ! -f "$PRISM" ]; then
    echo "❌ Error: prism binary not found at $PRISM"
    echo "   Run 'make build' first"
    exit 1
fi

echo "💰 Checking test instance costs (threshold: \$${THRESHOLD}/day)..."
echo ""

# Get workspace list output
WORKSPACE_OUTPUT=$($PRISM workspace list)

# Count test instances
TEST_COUNT=$(echo "$WORKSPACE_OUTPUT" | grep -c "^test-" || true)

if [ "$TEST_COUNT" -eq 0 ]; then
    echo "✅ No test instances running"
    exit 0
fi

# Extract daily cost estimate (last line of output)
DAILY_COST=$(echo "$WORKSPACE_OUTPUT" | grep "Estimated daily:" | awk '{print $3}' | tr -d '$')

if [ -z "$DAILY_COST" ]; then
    echo "⚠️  Warning: Could not parse daily cost"
    exit 0
fi

echo "📊 Cost Summary:"
echo "   Test instances: $TEST_COUNT"
echo "   Total daily: \$${DAILY_COST}"
echo ""

# Check threshold (using bc for float comparison)
if command -v bc >/dev/null 2>&1; then
    OVER_THRESHOLD=$(echo "$DAILY_COST > $THRESHOLD" | bc -l)

    if [ "$OVER_THRESHOLD" -eq 1 ]; then
        echo "🚨 ALERT: Daily cost (\$${DAILY_COST}) exceeds threshold (\$${THRESHOLD})"
        echo ""
        echo "   Recommended action:"
        echo "   ./scripts/cleanup-current-tests.sh"
        echo ""
        exit 1
    fi
fi

echo "✅ Daily cost within threshold"
echo ""
echo "   To clean up test instances:"
echo "   ./scripts/cleanup-current-tests.sh"
