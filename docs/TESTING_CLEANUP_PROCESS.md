# Test Instance Cleanup Process

## Overview
This document establishes a systematic process for managing test instances during development to prevent cost accumulation and maintain a clean environment.

## Principles

### 1. Clean As You Go
- **Delete test instances immediately after verification** - Don't let them accumulate
- **Document what was verified** before deletion (logs, screenshots, notes)
- **Use descriptive names** with purpose and date for tracking

### 2. Time-based Cleanup
- **Immediate** (0-1 hour): Instances used for quick feature tests
- **Short-term** (1-4 hours): Instances used for bug reproduction and verification
- **Medium-term** (4-24 hours): Instances used for integration testing
- **Long-term** (1+ days): Demo instances, persistent test environments

### 3. Cost Awareness
- **Check daily costs**: `prism workspace list` shows accumulated and estimated daily costs
- **Set cleanup reminders**: For long-running tests, set calendar reminders
- **Monitor per-instance cost**: Instances costing >$1/day should be reviewed

## Standard Testing Workflow

### Before Testing
```bash
# 1. Plan your test with a descriptive name
INSTANCE_NAME="test-feature-description-$(date +%Y%m%d)"

# 2. Document the test purpose in your notes
echo "Testing: Feature X fix for Issue #123" > /tmp/test-$INSTANCE_NAME.md

# 3. Launch the instance
prism launch template-name $INSTANCE_NAME --size M
```

### During Testing
```bash
# 1. Capture verification evidence BEFORE deleting
prism workspace exec $INSTANCE_NAME 'command-to-verify' | tee /tmp/test-$INSTANCE_NAME-output.txt

# 2. Document success/failure in your notes
echo "✅ Verified: Feature works as expected" >> /tmp/test-$INSTANCE_NAME.md

# 3. Check cloud-init logs if needed
prism workspace exec $INSTANCE_NAME 'cat /var/log/cloud-init-output.log' > /tmp/test-$INSTANCE_NAME-cloudinit.log
```

### After Testing
```bash
# 1. Verify the test objective was met
cat /tmp/test-$INSTANCE_NAME.md

# 2. Delete the instance immediately
prism workspace delete $INSTANCE_NAME

# 3. Move test notes to permanent location if needed
mv /tmp/test-$INSTANCE_NAME.md ~/test-results/$(date +%Y-%m)/
```

## Naming Conventions

### Test Instance Names
Use clear, structured names that indicate purpose and timing:

```bash
# ✅ GOOD: Descriptive with purpose
test-param-substitution-20260118
test-docker-inheritance-bug422
test-rstudio-latest-verification

# ❌ BAD: Vague or numbered sequentially
test-1
test-instance
my-test
```

### Name Format
```
test-<feature>-<purpose>[-<bugnum>][-<date>]

Examples:
- test-param-substitution-20260118
- test-conda-syspackages-bug422
- test-docker-foundation-verify
```

## Batch Cleanup Scripts

### Daily Cleanup Script
Create `scripts/cleanup-test-instances.sh`:

```bash
#!/bin/bash
# Daily cleanup of test instances

PRISM="./bin/prism"
DATE_THRESHOLD=$(date -v-1d +%Y-%m-%d)  # macOS
# DATE_THRESHOLD=$(date -d '1 day ago' +%Y-%m-%d)  # Linux

echo "🧹 Finding test instances older than $DATE_THRESHOLD..."

# List all test-* instances
$PRISM workspace list | grep "^test-" | while read -r line; do
    NAME=$(echo "$line" | awk '{print $1}')
    LAUNCHED=$(echo "$line" | awk '{print $9}')

    if [[ "$LAUNCHED" < "$DATE_THRESHOLD" ]]; then
        echo "Deleting old test instance: $NAME (launched $LAUNCHED)"
        $PRISM workspace delete "$NAME"
    fi
done
```

### End-of-Session Cleanup
Create `scripts/cleanup-current-tests.sh`:

```bash
#!/bin/bash
# Clean up all test-* instances from current session

PRISM="./bin/prism"

echo "🧹 Cleaning up all test-* instances..."

$PRISM workspace list | grep "^test-" | awk '{print $1}' | while read -r name; do
    echo "Deleting: $name"
    $PRISM workspace delete "$name"
done
```

## Integration with Development Workflow

### Git Pre-Push Hook
Add to `.git/hooks/pre-push`:

```bash
#!/bin/bash
# Remind about test instance cleanup before pushing

TEST_COUNT=$(./bin/prism workspace list | grep -c "^test-")

if [ "$TEST_COUNT" -gt 5 ]; then
    echo "⚠️  Warning: You have $TEST_COUNT test instances running"
    echo "   Consider cleaning up with: scripts/cleanup-current-tests.sh"
    echo ""
    read -p "Continue with push? (y/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi
```

### End-of-Day Checklist
Before ending a development session:

```bash
# 1. Check current test instances
prism workspace list | grep "^test-"

# 2. Review and clean up
scripts/cleanup-current-tests.sh

# 3. Verify demo instances are still needed
prism workspace list | grep -E "demo|collab|research"

# 4. Check daily cost estimate
prism workspace list | tail -5
```

## Emergency Cleanup

If you discover many orphaned test instances:

```bash
# 1. Export current instance list
prism workspace list > /tmp/instances-before-cleanup.txt

# 2. Create cleanup script from the list
grep "^test-" /tmp/instances-before-cleanup.txt | \
    awk '{print "prism workspace delete " $1}' > /tmp/cleanup-batch.sh

# 3. Review the script
cat /tmp/cleanup-batch.sh

# 4. Execute cleanup
chmod +x /tmp/cleanup-batch.sh
./bin/prism workspace list | wc -l  # Before count
/tmp/cleanup-batch.sh
./bin/prism workspace list | wc -l  # After count
```

## Monitoring and Alerts

### Cost Monitoring Script
Create `scripts/check-test-costs.sh`:

```bash
#!/bin/bash
# Alert if test instance costs exceed threshold

PRISM="./bin/prism"
THRESHOLD=50  # dollars per day

DAILY_COST=$($PRISM workspace list | grep "Estimated daily:" | awk '{print $3}' | tr -d '$')

if (( $(echo "$DAILY_COST > $THRESHOLD" | bc -l) )); then
    echo "🚨 ALERT: Daily cost is $${DAILY_COST} (threshold: $${THRESHOLD})"
    echo "   Run cleanup: scripts/cleanup-current-tests.sh"
    exit 1
fi

echo "✅ Daily cost: $${DAILY_COST} (within threshold)"
```

### Cron Job for Daily Cleanup
Add to crontab:

```bash
# Clean up test instances older than 1 day, every day at 2 AM
0 2 * * * cd /path/to/prism && scripts/cleanup-test-instances.sh >> logs/cleanup.log 2>&1
```

## Best Practices

### DO ✅
- Delete instances immediately after verification
- Capture logs and output BEFORE deletion
- Use descriptive instance names with dates
- Document what was verified
- Set time limits for long-running tests
- Check daily costs regularly

### DON'T ❌
- Leave test instances running overnight without justification
- Use vague names like "test-1", "my-instance"
- Forget to document what was tested
- Accumulate >10 test instances without cleanup
- Ignore cost warnings

## Phase-Based Cleanup

### Phase 1 Testing Example (This Session)
```bash
# Testing completed three bugs:
# 1. Parameter processor (test-simple, test-bash-latest, test-params-*)
# 2. Conda system packages (test-ds-*, test-docker-*)
# 3. Conda usermod newlines (test-ds-final verified)

# Created cleanup script for all Phase 1 instances:
/tmp/cleanup-phase1-instances.sh

# Result: 25 instances deleted, $50-100/day saved
```

### Future Phase Testing
For each testing phase:

1. **Create phase-specific cleanup script**
   ```bash
   cat > /tmp/cleanup-phase2-instances.sh
   # Add all phase 2 test instance names
   ```

2. **Run cleanup at phase completion**
   ```bash
   chmod +x /tmp/cleanup-phase2-instances.sh
   /tmp/cleanup-phase2-instances.sh
   ```

3. **Document phase results BEFORE cleanup**
   - What was tested
   - What was verified
   - What issues were found
   - Evidence captured

## Instance Lifecycle Summary

```
Launch → Test → Verify → Document → Delete
  ↓       ↓       ↓         ↓          ↓
 <1min   <30min  <5min    <2min     <1min

Total time per test: ~40 minutes
Cleanup: Immediate (don't wait!)
```

## Cost Impact

### Before Process
- Phase 1 testing: 60 instances accumulated
- Estimated daily cost: $257.80
- Cleanup done manually after completion

### With Process
- Test instances: <5 active at any time
- Estimated daily cost from tests: <$10
- Continuous cleanup prevents accumulation

### Savings
- **$240+/day saved** by cleaning as you go
- **$88,000+/year** if this becomes routine practice

## Related Documentation
- [Testing Guide](TESTING.md) - Comprehensive testing strategy
- [Phase 1 Testing Summary](/tmp/phase1-testing-summary.md) - Real example
- [Instance Cleanup Plan](/tmp/instance-cleanup-plan.md) - Analysis template
