# Prism Autonomous Idle Detection System

> **Updated in v0.35.0**: Idle detection has moved fully on-instance via the [spored daemon](SPORED_DAEMON.md). The previous daemon-side CloudWatch polling approach has been retired. This document reflects the current architecture.

## Overview

Prism automatically detects when a workspace is idle and hibernates or stops it to save costs. As of v0.35.0, all idle detection runs on the instance itself via the **spored daemon** — a lightweight Go service installed automatically on every workspace at launch.

## Architecture

```
┌─────────────────────────────────────────────────┐
│                   EC2 Instance                   │
│                                                  │
│  ┌────────────────────────────────────────────┐  │
│  │              spored daemon                 │  │
│  │                                            │  │
│  │  Reads prism:* EC2 tags for config         │  │
│  │  Monitors 7 local signals every 60s        │  │
│  │  Executes stop/hibernate when idle         │  │
│  └────────────────────────────────────────────┘  │
│                                                  │
└─────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────┐
│               Prism Daemon (local)               │
│                                                  │
│  Fleet policy management (EC2 tag writes)        │
│  TTL safety valve (stops expired instances)      │
│  Idle policy CRUD API                            │
└─────────────────────────────────────────────────┘
```

## Detection Signals

spored monitors 7 independent signals on the instance. All must be below threshold simultaneously for the idle timeout to trigger:

| Signal | Source | Default threshold |
|--------|--------|-----------------|
| CPU utilization | `/proc/stat` | < 5% |
| Network throughput | `/proc/net/dev` | < 10 KB/min |
| Disk I/O | `/proc/diskstats` | < 100 KB/min |
| GPU utilization | `nvidia-smi` (if present) | < 5% |
| Active terminal sessions | `/dev/pts/` directory | None |
| Logged-in users | `who` command | None |
| Recent user activity | `last` / `wtmp` | No activity in window |

This 7-signal approach is more accurate and cheaper than CloudWatch polling — it reads local `/proc` files instead of making AWS API calls.

## Configuration via EC2 Tags

spored reads configuration from `prism:*` EC2 instance tags at startup (and on `reload`):

| Tag | Example | Description |
|-----|---------|-------------|
| `prism:idle-timeout` | `1h` | Duration of sustained idleness before action |
| `prism:hibernate-on-idle` | `true` | Hibernate instead of stop |
| `prism:on-complete` | `stop` | Action when completion signal received |
| `prism:pre-stop` | `/home/ubuntu/cleanup.sh` | Shell command run before shutdown |

### Applying idle policies

Idle policies in the Prism GUI/CLI set these tags on the target instance. Policies are directives, not executors — spored reads the tags and acts on them.

```bash
prism admin idle-policy apply my-project --policy balanced
# → Sets prism:idle-timeout=1h, prism:hibernate-on-idle=true on instance tags
```

## Features

### ✅ On-Instance Detection
- **Accurate**: reads local system state, not AWS metrics (no CloudWatch delay)
- **Cheap**: no AWS API calls for idle checking
- **Responsive**: checks every 60 seconds; acts within 1 minute of threshold breach

### ✅ Intelligent Actions
- **Warnings**: broadcasts via `wall` to all logged-in users 5 minutes before action
- **Hibernation first**: preserves RAM state when instance type supports it
- **Auto-fallback**: stops instance if hibernation is unavailable
- **Pre-stop hooks**: run cleanup scripts before shutdown

### ✅ Fleet Policy Management
- Named policies (aggressive, balanced, conservative, research, development)
- Policies configured in GUI under Settings → Idle Detection
- Applied via EC2 tag writes — no agent restart required

## Implementation Details

### 1. IAM Role Setup

The system automatically creates and attaches the `Prism-Instance-Profile` IAM role to all launched instances:

**Permissions:**
```json
{
  "Version": "2012-10-17", 
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:CreateTags",
        "ec2:DescribeTags", 
        "ec2:DescribeInstances",
        "ec2:StopInstances"
      ],
      "Resource": "*"
    }
  ]
}
```

### 2. Template Integration

Templates with idle detection use the `idle_detection` configuration block:

```yaml
idle_detection:
  enabled: true
  idle_threshold_minutes: 5      # Start monitoring after 5 min idle
  hibernate_threshold_minutes: 10   # Hibernate/stop after 10 min idle  
  check_interval_minutes: 2      # Check every 2 minutes
```

### 3. Agent Deployment Process

**During Instance Launch:**
1. UserData script runs automatically
2. System packages updated (`apt-get update -y`)
3. Dependencies installed (`curl`, `unzip`, `bc`)
4. AWS CLI v2 installed (architecture-specific)
5. Idle detection agent script deployed
6. Cron job configured for periodic execution
7. Initial run scheduled after 2-minute delay

### 4. Agent Operation Flow

**Every Check Interval (default: 2 minutes):**

```bash
1. Get Instance Metadata (IMDSv2)
   ├── Instance ID via metadata service
   ├── Region detection
   └── Token-based authentication

2. Check System Activity
   ├── CPU load analysis  
   ├── Active user sessions
   ├── GPU utilization (if available)
   └── Network activity patterns

3. State Management
   ├── Read existing idle status from AWS tags
   ├── Calculate idle duration
   └── Apply progressive action thresholds

4. Action Execution
   ├── First Idle: Set idle timestamp
   ├── Sustained Idle (5+ min): Continue monitoring  
   ├── Hibernation Threshold (10+ min): Check hibernation support
   ├── Hibernation Available: aws ec2 stop-instances --hibernate
   └── Hibernation Unavailable: aws ec2 stop-instances
```

### 5. AWS Tag Schema

The system uses standardized AWS tags for state tracking:

| Tag | Values | Description |
|-----|--------|-------------|
| `Prism:IdleStatus` | `active`, `idle` | Current activity state |
| `Prism:IdleSince` | ISO8601 timestamp | When idle period started |
| `Prism:IdleAction` | `hibernating`, `hibernated`, `stopping`, `stopped` | Action taken |

## Daemon Integration

### Multi-Stage Detection

The daemon performs intelligent idle detection with three stages:

**Stage 1: Fast Rejection** (< 1 second)
- Active user connections via SSH
- Immediate "not idle" classification

**Stage 2: Research Work Detection**  
- Background computation without user interaction
- CPU, GPU, memory, disk activity analysis
- Research job pattern recognition

**Stage 3: True Idle Verification**
- Sustained quiet period validation
- Cross-verification with instance tags
- Progressive action evaluation

### Sample Daemon Output

```
🔍 Starting intelligent idle detection cycle...
  🔍 Stage 1: Checking for active user connections...
  Found 3 running instances with public IPs
    → idle-test has no active user connections
    → final-system-test has no active user connections
  🔍 Stage 2: Checking if system is busy with any work...
    → idle-test has low system activity
    → final-system-test is busy with background work
🔍 1 instances doing background research work - marked as non-idle
🔍 Stage 3: Verifying sustained quiet period...
    → idle-test appears to be truly idle
🔍 1 instances are truly idle - evaluating for cost-saving actions
🔍 Intelligent idle detection complete
```

## Configuration

### Template Configuration

```yaml
name: "Research Environment with Idle Detection"
description: "Automatically hibernates after 10 minutes of inactivity"

idle_detection:
  enabled: true
  idle_threshold_minutes: 5       # Alert threshold
  hibernate_threshold_minutes: 10 # Action threshold  
  check_interval_minutes: 2       # Monitoring frequency

# Agent automatically installed via user_data
user_data: |
  #!/bin/bash
  # AWS CLI v2 installation and agent deployment
  # (Full script included in template)
```

### Daemon Configuration

The daemon automatically enables autonomous monitoring when:
- Idle detection is enabled in the idle manager
- Running instances are detected with public IPs
- Proper AWS profile configuration is available

## Version Management

### Agent Versioning

The idle detection agent includes version tracking:

```bash
# Agent version and metadata  
AGENT_VERSION="1.0.0"
AGENT_BUILD_DATE="2025-08-08"
MIN_AWS_CLI_VERSION="2.0.0"
```

**Version Logging:**
```
2025-08-08 16:48:00 [IDLE-AGENT v1.0.0] Prism Idle Detection Agent v1.0.0 (built 2025-08-08)
2025-08-08 16:48:00 [IDLE-AGENT v1.0.0] AWS CLI version: 2.28.5
```

## Testing and Validation

### End-to-End Test Results

**Test Instance: `final-system-test`**

✅ **Launch**: Template deployed successfully with idle detection
✅ **Installation**: AWS CLI v2 and agent installed automatically  
✅ **Detection**: System correctly identified as idle after user disconnect
✅ **Duration Tracking**: 9+ minute idle period accurately measured
✅ **Hibernation Check**: Detected instance doesn't support hibernation  
✅ **Fallback Action**: Automatically stopped instance (cost = $0.00)
✅ **State Tracking**: AWS tags properly maintained throughout lifecycle

### Validation Commands

```bash
# Check agent deployment
ssh ubuntu@<instance-ip> "ls -la /usr/local/bin/prism-idle-check.sh"
ssh ubuntu@<instance-ip> "/usr/local/bin/aws --version"  
ssh ubuntu@<instance-ip> "cat /etc/cron.d/prism-idle"

# Test agent execution
ssh ubuntu@<instance-ip> "sudo /usr/local/bin/prism-idle-check.sh"

# Verify AWS tags
aws ec2 describe-tags --filters "Name=resource-id,Values=<instance-id>" \
  --query 'Tags[?starts_with(Key, `Prism:`)].{Key:Key,Value:Value}' --output table

# Check daemon monitoring
tail -f daemon.log | grep "idle detection"
```

## Cost Savings Impact

### Before Idle Detection
- **Running Instances**: Continue consuming compute resources
- **Manual Management**: Users must remember to stop instances
- **Cost Leakage**: Forgotten instances accumulate charges

### After Idle Detection
- **Automatic Stopping**: Idle instances stopped within 10 minutes
- **Zero Compute Cost**: Stopped instances only pay for EBS storage  
- **Hibernation Support**: RAM state preserved when available
- **Audit Trail**: Complete tracking of all automated actions

### Example Savings
- **t3.medium instance**: $0.0464/hour × 24 hours = $1.11/day
- **With 10-minute idle detection**: Maximum waste = $0.008/day
- **Daily savings**: ~$1.10 per idle instance
- **Monthly savings**: ~$33 per idle instance

## Troubleshooting

### Common Issues

1. **Agent Not Installed**
   - Check UserData execution: `sudo cat /var/log/cloud-init-output.log`
   - Verify template deployment: Check if custom UserData is being used

2. **AWS CLI Missing** 
   - Check installation logs: `sudo tail -100 /var/log/cloud-init-output.log`
   - Architecture detection: `uname -m` should match downloaded CLI

3. **Permission Errors**
   - Verify IAM role: `aws sts get-caller-identity`
   - Check instance profile: `curl http://169.254.169.254/latest/meta-data/iam/security-credentials/`

4. **IMDSv2 Token Errors**
   - Test token generation: `curl -X PUT "http://169.254.169.254/latest/api/token" -H "X-aws-ec2-metadata-token-ttl-seconds: 21600"`
   - Verify metadata access with token

### Debug Commands

```bash
# Check agent status
sudo /usr/local/bin/prism-idle-check.sh

# View agent logs  
tail -50 /var/log/prism-idle.log

# Test cron job
sudo run-parts --test /etc/cron.d/

# Verify AWS permissions
aws ec2 describe-tags --filters "Name=resource-id,Values=$(curl -s http://169.254.169.254/latest/meta-data/instance-id)"
```

## Future Enhancements (TODOs)

### High Priority
- [ ] **Template Software Install Testing**: Validate that normal templates deploy software correctly
- [ ] **Agent Update Mechanism**: Automatic agent updates when new versions available
- [ ] **AWS CLI Update Checking**: Periodic AWS CLI updates pushed to instances
- [ ] **Enhanced Metrics**: GPU utilization, network I/O, disk activity monitoring

### Medium Priority  
- [ ] **Hibernation Policy Optimization**: Machine learning-based hibernation vs stop decisions
- [ ] **Multi-User Support**: Per-user idle detection and notification
- [ ] **Research Domain Intelligence**: Domain-specific idle patterns (ML training, data analysis)
- [ ] **Cost Analytics Integration**: Real-time savings reporting

### Low Priority
- [ ] **Web UI Dashboard**: Visual monitoring of idle detection across all instances  
- [ ] **Slack/Email Notifications**: Alerts before automatic actions
- [ ] **Custom Threshold Profiles**: Research group-specific idle policies
- [ ] **Integration Testing**: Automated end-to-end validation suite

---

*This system represents a major advancement in Prism's cost optimization capabilities, providing researchers with automatic instance management while preserving work state and minimizing compute waste.*