# Prism Troubleshooting Guide

## Quick Fixes for Common Issues

### 🚨 "daemon not running" Error

**What you see:**
```
Error: daemon not running
```

**Quick fix:**
```bash
# Daemon usually auto-starts — try your command again
# If it's failing, check status:
prism admin daemon status

# Stop and let it auto-restart on next command
prism admin daemon stop
prism templates
```

**If daemon won't start:**
```bash
# Check if something is using port 8947
lsof -i :8947

# Kill conflicting process if found
kill -9 <PID>

# Try running a prism command to trigger auto-start
prism templates
```

---

### 🔐 AWS Credential Issues

**What you see:**
```
Error: AWS credentials not found
Error: UnauthorizedOperation
```

**Quick fix:**
```bash
# Check current credentials
aws sts get-caller-identity

# Configure if needed
aws configure
```

**If you have AWS credentials but Prism can't find them:**
```bash
# Check AWS profile
echo $AWS_PROFILE

# Set profile if needed
export AWS_PROFILE=your-profile-name

# Or specify directly
prism workspace launch python-ml my-project --profile your-profile-name
```

---

### 🏗️ Template Launch Failures

**What you see:**
```
Error: failed to launch instance
Error: VPC not found
Error: subnet not available
```

**Quick fix:**
```bash
# Prism auto-discovers VPC/subnet
prism workspace launch python-ml my-project

# If auto-discovery fails, check your VPC setup
aws ec2 describe-vpcs --query 'Vpcs[?IsDefault==`true`]'
```

**If you don't have a default VPC:**
```bash
# Create a default VPC
aws ec2 create-default-vpc
```

---

### 💰 Cost and Pricing Concerns

**What you see:**
```
Instance cost seems high
Unexpected AWS charges
```

**Quick fix:**
```bash
# Check current instances and costs
prism workspace list

# Stop unused instances
prism workspace stop instance-name

# Use hibernation to preserve work and reduce costs
prism workspace hibernate instance-name
```

**Cost optimization commands:**
```bash
# Use smaller instance sizes
prism workspace launch python-ml my-project --size S

# Use spot instances (up to 90% savings)
prism workspace launch python-ml my-project --spot
```

---

### 🔌 Connection Problems

**What you see:**
```
Connection timeout
SSH connection refused
Can't access Jupyter/RStudio
```

**Quick fix:**
```bash
# Check instance status
prism workspace list

# Ensure instance is running
prism workspace resume instance-name

# Get fresh connection info
prism workspace connect instance-name
```

**If SSH still fails:**
```bash
# Check instance status details
prism workspace list

# Wait for instance to fully boot (can take 2-3 minutes)
# Then try connecting again
```

---

### 🧠 Memory and Performance Issues

**What you see:**
```
Instance running slowly
Out of memory errors
Jupyter kernel crashes
```

**Quick fix:**
```bash
# Delete and relaunch with larger instance size
prism workspace delete instance-name
prism workspace launch python-ml instance-name --size L

# Or add more EFS storage for data
prism volume create extra-space
prism volume attach extra-space instance-name
```

---

### 📦 Template and Package Issues

**What you see:**
```
Package not found
Template validation failed
Command not available in instance
```

**Quick fix:**
```bash
# Validate template before launching
prism templates validate python-ml

# Check template contents
prism templates info python-ml
```

**If template seems broken:**
```bash
# Force refresh template cache
rm -rf ~/.prism/templates
prism templates
```

---

### 🌍 Region and Availability Issues

**What you see:**
```
Insufficient capacity
Instance type not available
AMI not found in region
```

**Quick fix:**
```bash
# Try different region
prism workspace launch python-ml my-project --region us-east-1

# Use different instance size
prism workspace launch python-ml my-project --size M

# Check region availability
aws ec2 describe-availability-zones --region us-west-2
```

---

### 🔧 GUI and Interface Issues

**What you see:**
```
GUI won't start
TUI looks broken
Interface unresponsive
```

**Quick fix:**
```bash
# For TUI display issues
export TERM=xterm-256color
prism tui

# For interface problems, use CLI as backup
prism workspace list
prism workspace connect instance-name
```

---

## Advanced Troubleshooting

### Enable Debug Logging
```bash
# Set debug mode
export PRISM_DEBUG=1

# Check daemon status
prism admin daemon status

# Or run commands with verbose output
prism workspace launch python-ml my-project --verbose
```

### Check System Requirements
```bash
# Verify AWS CLI version (need v2+)
aws --version

# Check Prism version
prism version

# Verify network connectivity
curl -I https://ec2.amazonaws.com
```

### Reset Prism
```bash
# Stop daemon
prism admin daemon stop

# Clear cache and state
rm -rf ~/.prism/

# Restart fresh (daemon auto-starts on next command)
prism templates
```

---

## Getting Help

### Before Opening an Issue

1. **Check daemon status**: `prism admin daemon status`
2. **Verify AWS credentials**: `aws sts get-caller-identity`
3. **Try CLI interface**: Sometimes GUI/TUI have display issues
4. **Check recent changes**: Did you update AWS credentials or change regions?

### Include This Information

When asking for help, please include:

```bash
# Prism version
prism version

# Daemon status
prism admin daemon status

# AWS account info (no credentials)
aws sts get-caller-identity --query 'Account'

# Operating system
uname -a

# Error message (full text)
```

### Community Support

- **GitHub Issues**: [Report bugs and request features](https://github.com/scttfrdmn/prism/issues)
- **Discussions**: [Get help from the community](https://github.com/scttfrdmn/prism/discussions)
- **Documentation**: [Complete guides in `/docs` folder](docs/)

---

## Emergency Recovery

### Instance Stuck in Bad State
```bash
# Force stop
prism workspace stop instance-name

# Delete and recreate
prism ami save instance-name "backup-before-delete"   # Save work first if possible
prism workspace delete instance-name
prism workspace launch --ami "backup-before-delete" instance-name-new
```

### Accidentally Deleted Important Instance
```bash
# Check for saved AMIs
prism ami list

# Contact AWS support for EBS snapshot recovery if critical
```

### Unexpected High AWS Bills
```bash
# Immediately stop all instances
prism workspace list
# Stop each running instance:
prism workspace stop instance-name-1
prism workspace stop instance-name-2

# Review hibernation options for the future
# Templates include idle detection - check template settings
```

**Remember**: Prism is designed to "default to success." Most issues have simple solutions, and the error messages are designed to guide you to the fix.
