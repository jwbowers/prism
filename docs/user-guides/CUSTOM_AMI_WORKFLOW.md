# Custom AMI Workflow Guide

Complete guide to creating, managing, and sharing custom Amazon Machine Images (AMIs) with Prism for fast instance provisioning.

## Table of Contents

- [Overview](#overview)
- [Why Use Custom AMIs?](#why-use-custom-amis)
- [Quick Start](#quick-start)
- [Complete Workflow](#complete-workflow)
- [Command Reference](#command-reference)
- [Use Cases](#use-cases)
- [Performance Comparison](#performance-comparison)
- [Troubleshooting](#troubleshooting)

## Overview

Custom AMIs are pre-configured instance snapshots that enable:
- **Fast Launching**: 30-60 seconds vs 3-5 minutes for templates
- **Consistent Environments**: Everyone uses identical configuration
- **Rapid Iteration**: Quick launch → test → terminate cycles
- **Workshop Ready**: Provision dozens of identical instances instantly

Perfect for research teams, teaching, workshops, and production deployments.

## Why Use Custom AMIs?

### The Problem
Templates install packages every time an instance launches:
```bash
# Every template launch repeats the same slow steps:
# 1. Boot base OS (1 min)
# 2. Install R/Python (30 sec)
# 3. Install packages (2-4 min) ← SLOW!
# Total: 3-5 minutes per instance
```

### The Solution
Create an AMI once, launch instantly:
```bash
# One-time setup:
prism ami create my-configured-env --name "My Research Environment"
# Takes 5-10 minutes (one time)

# Every subsequent launch:
prism launch --ami ami-abc123def456 quick-instance
# Takes 30-60 seconds ← FAST!
```

**Result**: 83% faster launches after initial AMI creation.

## Quick Start

### 5-Minute AMI Creation

```bash
# 1. Launch a base template
prism launch r-research my-r-env

# 2. Connect and customize
prism connect my-r-env
# Install your packages, configure settings, etc.
sudo R -e 'install.packages(c("tidyverse", "caret", "randomForest"))'
exit

# 3. Create AMI from configured instance
prism ami create my-r-env \
  --name "Team R Environment v1.0" \
  --description "R 4.4.2 + RStudio + ML packages"

# 4. Launch from your custom AMI (30 seconds!)
prism launch --ami ami-abc123def456 quick-start
```

## Complete Workflow

### Step 1: Launch and Customize Base Template

Start with a Prism template that's close to your needs:

```bash
# Launch base R research environment
prism launch r-research my-base-env --size M

# Wait for launch to complete
prism status my-base-env
# State: running

# Connect to instance
prism connect my-base-env
```

Inside the instance, customize your environment:

```bash
# Install additional R packages
sudo R
> install.packages(c(
+   "tidymodels",     # Modern ML framework
+   "arrow",          # Fast data I/O
+   "targets",        # Workflow management
+   "renv"            # Package versioning
+ ))
> q()

# Install system dependencies
sudo apt-get update
sudo apt-get install -y libgdal-dev libproj-dev

# Configure RStudio settings
echo 'options(repos = "https://cran.rstudio.com/")' >> ~/.Rprofile

# Add custom scripts or data
mkdir -p ~/scripts
cat > ~/scripts/setup.R <<'EOF'
# Team-specific R setup
library(tidyverse)
library(tidymodels)
message("Team R Environment v1.0 loaded")
EOF

# Exit when done
exit
```

### Step 2: Create Custom AMI

Convert your configured instance into an AMI:

```bash
# Create AMI with descriptive name and tags
prism ami create my-base-env \
  --name "Team R Environment v1.0" \
  --description "R 4.4.2 + RStudio + tidymodels + arrow + targets + renv" \
  --tags "team=research,project=climate-analysis,version=1.0,environment=production"

# Output:
# Creating AMI from instance my-base-env...
# AMI ID: ami-0abc123def456789
# Creation started. This takes 5-10 minutes.
# Use 'prism ami list' to check progress.
```

**Important**: AMI creation does NOT stop your instance. You can continue working.

Check creation progress:

```bash
# List your AMIs with status
prism ami list --owner self

# Output:
# AMI ID              Name                       State      Created
# ami-0abc123def456   Team R Environment v1.0    pending    2026-01-16 10:30
```

Wait for state to become `available` (5-10 minutes).

View detailed AMI information:

```bash
# Get complete AMI details
prism ami describe ami-0abc123def456

# Output shows:
# - Current state (pending/available/failed)
# - Size and storage details
# - Tags and metadata
# - Creation timestamp
```

### Step 3: Launch Instances from Your AMI

Once AMI is available, launch instances in ~30 seconds:

```bash
# Launch single instance from AMI
prism launch --ami ami-0abc123def456 researcher-1

# Launch multiple instances (workshop scenario)
for i in {1..20}; do
  prism launch --ami ami-0abc123def456 workshop-instance-$i --size S &
done
wait

# All 20 instances launch in parallel, ready in ~1 minute total!
```

Launch with custom size and spot pricing:

```bash
# Launch with specific instance type
prism launch --ami ami-0abc123def456 analysis-job \
  --size XL \
  --spot \
  --project climate-research
```

Verify your custom environment is intact:

```bash
# Connect and test
prism connect researcher-1

# Check that your packages and scripts are present
ls ~/scripts/
R --version
Rscript ~/scripts/setup.R
```

### Step 4: Share AMI with Team

Make your AMI available to team members:

#### Share with Specific AWS Accounts

```bash
# Share with specific AWS account IDs
prism ami share ami-0abc123def456 \
  --accounts 123456789012,987654321098 \
  --regions us-west-2,us-east-1

# Output:
# Sharing AMI ami-0abc123def456...
# ✓ Shared with account 123456789012 in us-west-2
# ✓ Shared with account 123456789012 in us-east-1
# ✓ Shared with account 987654321098 in us-west-2
# ✓ Shared with account 987654321098 in us-east-1
```

Team members can now launch using the shared AMI ID:

```bash
# Team member in another AWS account
prism launch --ami ami-0abc123def456 my-instance
# Works! They get the same environment.
```

#### Publish to Prism Marketplace (Public)

Make your AMI available to the entire Prism community:

```bash
# Publish to public marketplace
prism ami publish ami-0abc123def456 \
  --name "Climate Research R Environment" \
  --category "Research Tools" \
  --description "Complete R environment for climate data analysis with tidyverse, tidymodels, arrow, and spatial packages" \
  --readme ./AMI_README.md

# Output:
# Publishing AMI to Prism marketplace...
# ✓ AMI published successfully
# ✓ Marketplace listing created
#
# Share this command with users:
#   prism launch --ami ami-0abc123def456 my-instance
#
# View in marketplace:
#   prism marketplace browse --category "Research Tools"
```

### Step 5: Version Management

Track and manage AMI versions over time:

#### Tag for Version Control

```bash
# Add semantic version tags
prism ami tag ami-0abc123def456 \
  --tags "version=1.0.0,status=production,date=2026-01-16"

# Update version when you create a new AMI
prism ami create my-base-env-v2 \
  --name "Team R Environment v1.1" \
  --tags "version=1.1.0,status=production,previous=ami-0abc123def456"
```

#### Deprecate Old Versions

```bash
# Mark old version as deprecated
prism ami deprecate ami-0abc123def456 \
  --deprecation-date 2026-03-01 \
  --replacement ami-0xyz987fed654321

# Output:
# AMI ami-0abc123def456 marked as deprecated.
# Will be deregistered on 2026-03-01.
# Recommended replacement: ami-0xyz987fed654321
```

Users launching the deprecated AMI will see a warning:

```bash
prism launch --ami ami-0abc123def456 my-instance
# ⚠️  Warning: AMI ami-0abc123def456 is deprecated
# Deprecated on: 2026-01-16
# Will be removed: 2026-03-01
# Recommended replacement: ami-0xyz987fed654321
# Continue anyway? [y/N]
```

#### Clean Up Old AMIs

```bash
# List all your AMIs
prism ami list --owner self

# Delete AMIs you no longer need (saves storage costs)
prism ami delete ami-old123def456

# Output:
# Are you sure you want to delete AMI ami-old123def456?
# This cannot be undone. [y/N] y
# ✓ AMI ami-old123def456 deleted
# ✓ Snapshots deleted
# Freed: 8.5 GB storage
```

### Step 6: Copy to Multiple Regions

For multi-region teams or global workshops:

```bash
# Copy AMI to other regions
prism ami copy ami-0abc123def456 \
  --source-region us-west-2 \
  --regions us-east-1,eu-west-1,ap-southeast-1 \
  --name "Team R Environment v1.0"

# Output:
# Copying AMI to 3 regions...
# ✓ us-east-1: ami-0new123def456 (copying, ~5 min)
# ✓ eu-west-1: ami-0new789abc123 (copying, ~5 min)
# ✓ ap-southeast-1: ami-0new456fed789 (copying, ~5 min)
```

Check copy progress:

```bash
# List AMIs in specific region
prism ami list --region us-east-1 --owner self
```

## Command Reference

### Core AMI Commands

#### `prism ami create`

Create an AMI from a running or stopped instance.

```bash
prism ami create INSTANCE_NAME [flags]

Flags:
  --name string          AMI name (required)
  --description string   AMI description
  --tags key=value       Tags (comma-separated)
  --no-reboot           Don't reboot instance during creation (faster but less reliable)
```

**Examples**:

```bash
# Basic AMI creation
prism ami create my-instance --name "My Custom AMI"

# With full metadata
prism ami create my-instance \
  --name "Production R Environment v2.0" \
  --description "R 4.4.2 + RStudio + Production packages" \
  --tags "version=2.0,team=research,environment=prod"

# Fast creation (no reboot, slightly less reliable)
prism ami create my-instance \
  --name "Quick AMI" \
  --no-reboot
```

#### `prism ami list`

List available AMIs.

```bash
prism ami list [flags]

Flags:
  --owner string        Filter by owner (self, amazon, aws-marketplace, or AWS account ID)
  --region string       AWS region (default: current profile region)
  --name-filter string  Filter by name pattern
  --tags key=value      Filter by tags
```

**Examples**:

```bash
# List your AMIs
prism ami list --owner self

# List AMIs in specific region
prism ami list --owner self --region us-east-1

# Filter by name
prism ami list --owner self --name-filter "R Environment*"

# Filter by tags
prism ami list --owner self --tags "team=research,version=1.*"
```

#### `prism ami describe`

Show detailed information about a specific AMI.

```bash
prism ami describe AMI_ID [flags]

Flags:
  --region string  AWS region (default: current profile region)
  --json           Output in JSON format
```

**Examples**:

```bash
# Basic describe
prism ami describe ami-0abc123def456

# JSON output for scripting
prism ami describe ami-0abc123def456 --json
```

#### `prism ami delete`

Delete an AMI and its associated snapshots.

```bash
prism ami delete AMI_ID [flags]

Flags:
  --region string  AWS region (default: current profile region)
  --force          Skip confirmation prompt
```

**Examples**:

```bash
# Interactive deletion (with confirmation)
prism ami delete ami-0abc123def456

# Force deletion (no prompt, for scripts)
prism ami delete ami-0abc123def456 --force
```

#### `prism ami share`

Share an AMI with other AWS accounts.

```bash
prism ami share AMI_ID [flags]

Flags:
  --accounts strings  AWS account IDs (comma-separated)
  --regions strings   Regions to share in (comma-separated)
  --public            Make AMI public (use with caution)
```

**Examples**:

```bash
# Share with specific accounts
prism ami share ami-0abc123def456 \
  --accounts 123456789012,987654321098

# Share across regions
prism ami share ami-0abc123def456 \
  --accounts 123456789012 \
  --regions us-west-2,us-east-1

# Make public (use with caution!)
prism ami share ami-0abc123def456 --public
```

#### `prism ami publish`

Publish an AMI to the Prism marketplace.

```bash
prism ami publish AMI_ID [flags]

Flags:
  --name string         Marketplace listing name (required)
  --category string     Category (required)
  --description string  Full description (required)
  --readme string       Path to README file
  --version string      Version number (semantic versioning recommended)
```

**Examples**:

```bash
# Publish to marketplace
prism ami publish ami-0abc123def456 \
  --name "Climate Research R Environment" \
  --category "Research Tools" \
  --description "Complete R environment for climate data analysis" \
  --readme ./AMI_README.md \
  --version "1.0.0"
```

#### `prism ami tag`

Add or update tags on an AMI.

```bash
prism ami tag AMI_ID [flags]

Flags:
  --tags key=value  Tags to add/update (comma-separated)
  --region string   AWS region (default: current profile region)
```

**Examples**:

```bash
# Add version tag
prism ami tag ami-0abc123def456 --tags "version=1.1.0"

# Update multiple tags
prism ami tag ami-0abc123def456 \
  --tags "status=production,tested=true,date=2026-01-16"
```

#### `prism ami deprecate`

Mark an AMI as deprecated.

```bash
prism ami deprecate AMI_ID [flags]

Flags:
  --deprecation-date string  Date to deprecate (YYYY-MM-DD)
  --replacement string       Recommended replacement AMI ID
```

**Examples**:

```bash
# Deprecate with replacement
prism ami deprecate ami-0abc123def456 \
  --deprecation-date 2026-03-01 \
  --replacement ami-0xyz987fed654321
```

#### `prism ami copy`

Copy an AMI to other regions.

```bash
prism ami copy AMI_ID [flags]

Flags:
  --source-region string  Source region (required)
  --regions strings       Destination regions (comma-separated, required)
  --name string          Name for copied AMI
```

**Examples**:

```bash
# Copy to multiple regions
prism ami copy ami-0abc123def456 \
  --source-region us-west-2 \
  --regions us-east-1,eu-west-1 \
  --name "Team R Environment v1.0"
```

### Launching from AMIs

#### `prism launch --ami`

Launch an instance from a custom AMI.

```bash
prism launch --ami AMI_ID INSTANCE_NAME [flags]

Flags:
  --size string      Instance size (XS/S/M/L/XL/XXL)
  --spot             Use spot pricing
  --project string   Project ID for cost tracking
  --region string    AWS region
```

**Examples**:

```bash
# Basic AMI launch
prism launch --ami ami-0abc123def456 my-instance

# With custom size and spot pricing
prism launch --ami ami-0abc123def456 analysis-job \
  --size XL \
  --spot \
  --project climate-research

# Launch in specific region
prism launch --ami ami-0abc123def456 eu-instance \
  --region eu-west-1
```

## Use Cases

### 1. Team Consistency

**Problem**: Team members have different package versions, causing "works on my machine" issues.

**Solution**: Everyone launches from the same AMI.

```bash
# Team lead creates standardized environment
prism ami create team-base --name "Team Environment 2026.1"

# Share with team
prism ami share ami-0abc123def456 --accounts 111,222,333

# Everyone launches identical environment
prism launch --ami ami-0abc123def456 my-work
```

**Result**: Zero configuration drift, consistent results across team.

### 2. Conference Workshops

**Problem**: Need to provision 40 identical instances for workshop attendees.

**Solution**: Pre-create AMI, launch all instances in minutes.

```bash
# Before workshop: Create AMI
prism ami create workshop-template \
  --name "Workshop Environment 2026" \
  --description "Includes datasets and pre-configured notebooks"

# During workshop: Launch for all attendees (parallel)
for i in {1..40}; do
  prism launch --ami ami-0abc123def456 workshop-attendee-$i --size S &
done
wait

# Result: 40 instances ready in ~2 minutes (vs 2+ hours with templates)
```

**Result**: Happy attendees, smooth workshop, minimal wait time.

### 3. Fast Iteration for Development

**Problem**: Need to test code changes repeatedly with clean environments.

**Solution**: Launch → test → terminate cycle in minutes.

```bash
# Create base testing AMI once
prism ami create test-base --name "Testing Environment"

# Fast iteration loop
while true; do
  # Launch test instance (30 seconds)
  prism launch --ami ami-0abc123def456 test-run

  # Run tests
  prism exec test-run "Rscript /path/to/tests.R"

  # Terminate
  prism delete test-run --force

  # Repeat!
done
```

**Result**: 10x faster development iteration.

### 4. Pre-installed Research Packages

**Problem**: Installing bioinformatics packages takes hours.

**Solution**: Install once in AMI, use forever.

```bash
# One-time: Create AMI with all packages
prism launch r-research bioinfo-base
prism connect bioinfo-base
# ... install all Bioconductor packages (2-3 hours)
exit

prism ami create bioinfo-base \
  --name "Bioinformatics Environment 2026"

# Every subsequent launch: 30 seconds!
prism launch --ami ami-0abc123def456 analysis-1
```

**Result**: Save hours per launch, happier researchers.

### 5. Backup and Restore

**Problem**: Need to save instance state before major changes.

**Solution**: Create AMI as snapshot before risky operations.

```bash
# Before dangerous operation
prism ami create my-instance \
  --name "Backup before migration $(date +%Y%m%d)"

# Perform risky operation
prism connect my-instance
# ... make changes ...

# If something breaks, restore from AMI
prism launch --ami ami-backup-20260116 my-instance-restored
```

**Result**: Confidence to experiment, easy rollback.

## Performance Comparison

### Time Savings

| Scenario | Template | AMI | Time Saved |
|----------|----------|-----|------------|
| **Single launch** | 3-5 min | 30-60 sec | **~4 min** (83% faster) |
| **10 instances** | 30-50 min | 5-10 min | **~40 min** (80% faster) |
| **40 instances (workshop)** | 2-3 hours | 2-5 min | **~2.8 hours** (96% faster) |
| **Daily dev iterations (10x)** | 30-50 min | 5-10 min | **~40 min/day** (80% faster) |

### Cost Analysis

#### Storage Costs

AMIs cost $0.05/GB/month for EBS snapshots:

```
Typical AMI sizes:
- Base OS: ~8 GB = $0.40/month
- R + RStudio: ~12 GB = $0.60/month
- Python ML: ~15 GB = $0.75/month
- Full Research Stack: ~20 GB = $1.00/month
```

#### Time Savings Value

If researcher time = $50/hour:

```
Time saved per 10 launches:
- 40 minutes saved
- Value: 40 min × ($50/60 min) = $33.33

AMI storage cost: $1.00/month
Monthly launches: ~50 (5/week)
Value generated: 50 × $3.33 = $166.50/month

ROI: $166.50 / $1.00 = 16,650% monthly ROI
```

**Conclusion**: AMIs pay for themselves immediately in time savings.

## Troubleshooting

### AMI Creation Fails

**Symptom**: AMI creation hangs or fails.

**Causes & Solutions**:

1. **Instance not in stable state**
   ```bash
   # Check instance state
   prism status my-instance
   # Must be 'running' or 'stopped'

   # Wait for instance to stabilize
   prism stop my-instance
   prism status my-instance  # Wait for 'stopped'
   prism ami create my-instance --name "My AMI"
   ```

2. **Insufficient EBS snapshot space**
   ```bash
   # Check your AWS EBS snapshot limits
   aws service-quotas get-service-quota \
     --service-code ebs \
     --quota-code L-309BACF6

   # Delete old snapshots if needed
   prism ami list --owner self
   prism ami delete ami-old123def456
   ```

3. **AMI creation timeout**
   ```bash
   # Large instances take longer (10-15 min)
   # Check AWS console for detailed error messages
   # https://console.aws.amazon.com/ec2/v2/home?region=us-west-2#Images:
   ```

### AMI Launch Fails

**Symptom**: Cannot launch instance from AMI.

**Causes & Solutions**:

1. **AMI not in correct region**
   ```bash
   # Check which region AMI is in
   prism ami describe ami-0abc123def456

   # Copy to desired region
   prism ami copy ami-0abc123def456 \
     --source-region us-west-2 \
     --regions us-east-1
   ```

2. **AMI not shared with account**
   ```bash
   # Verify AMI permissions
   prism ami describe ami-0abc123def456

   # If owned by another account, ask them to share
   prism ami share ami-0abc123def456 \
     --accounts YOUR_ACCOUNT_ID
   ```

3. **AMI deprecated or deleted**
   ```bash
   # Check AMI status
   prism ami describe ami-0abc123def456
   # State should be 'available', not 'pending' or 'failed'
   ```

### AMI Takes Too Long to Create

**Symptom**: AMI creation exceeds 15 minutes.

**Optimization**:

```bash
# 1. Stop instance first (faster creation)
prism stop my-instance
prism ami create my-instance --name "My AMI"

# 2. Use --no-reboot flag (faster but riskier)
prism ami create my-instance --name "My AMI" --no-reboot

# 3. Reduce instance disk size before creating AMI
prism connect my-instance
# Inside instance:
sudo apt-get clean
sudo rm -rf /tmp/*
sudo rm -rf ~/.cache/*
exit
prism ami create my-instance --name "Clean AMI"
```

### Shared AMI Not Accessible

**Symptom**: Team member cannot launch your shared AMI.

**Solution**:

```bash
# 1. Verify you shared with correct account ID
prism ami describe ami-0abc123def456
# Check "SharedWith" field

# 2. Verify region
# AMI must be in same region or copied
prism ami copy ami-0abc123def456 \
  --source-region us-west-2 \
  --regions us-east-1

# 3. Re-share if needed
prism ami share ami-0abc123def456 \
  --accounts TEAM_MEMBER_ACCOUNT_ID \
  --regions us-west-2,us-east-1
```

### High Storage Costs

**Symptom**: AWS bill shows high EBS snapshot charges.

**Solution**:

```bash
# 1. List all your AMIs
prism ami list --owner self

# 2. Delete unused AMIs
prism ami delete ami-old123def456
prism ami delete ami-old789abc012

# 3. Automate cleanup with script
#!/bin/bash
# Delete AMIs older than 90 days
prism ami list --owner self --json | \
  jq -r '.[] | select(.CreationDate < "2025-10-01") | .ImageId' | \
  while read ami_id; do
    prism ami delete $ami_id --force
  done
```

## Next Steps

- **Learn Best Practices**: See [AMI_BEST_PRACTICES.md](AMI_BEST_PRACTICES.md)
- **Browse Marketplace**: `prism marketplace browse`
- **Template Guide**: See [R_TEMPLATE_GUIDE.md](R_TEMPLATE_GUIDE.md) for combining templates with AMIs
- **Join Community**: Share your AMIs at https://github.com/scttfrdmn/prism/discussions

## Related Documentation

- [Multi-User Instance Setup](MULTI_USER_INSTANCE_SETUP.md) - Adding users to AMI-based instances
- [Cost Management](COST_MANAGEMENT.md) - Optimizing AMI storage costs
- [R Research Guide](R_TEMPLATE_GUIDE.md) - R-specific AMI creation tips
