# Getting Started with Prism

## Quick Start (5 minutes)

Prism provides pre-configured research environments without complex setup requirements.

### 1. Installation

See the main [Installation Guide](../index.md#installation) for detailed installation instructions for your platform (macOS, Linux, Windows, Conda).

Quick install:
```bash
# macOS/Linux
brew tap scttfrdmn/prism
brew install prism
```

```powershell
# Windows
scoop bucket add scttfrdmn https://github.com/scttfrdmn/scoop-bucket
scoop install prism
```

### 2. AWS Setup

Prism uses your existing AWS credentials. If you don't have AWS CLI configured:

```bash
aws configure
# Enter your AWS Access Key ID, Secret Access Key, and default region
```

For detailed AWS setup including IAM permissions, see the [Administrator Guide](../admin-guides/ADMINISTRATOR_GUIDE.md) or [AWS IAM Permissions](../admin-guides/AWS_IAM_PERMISSIONS.md).

### 3. Launch Your First Environment
```bash
# See available templates
prism templates

# Launch a Python ML environment
prism workspace launch python-ml my-first-project

# Get connection info
prism workspace connect my-first-project
```

That's it! Your research environment is ready.

---

## Choose Your Interface

Prism offers three ways to interact:

### 🖥️ **GUI (Desktop App)**
Perfect for visual management and one-click operations.
```bash
prism gui
```

### 📱 **TUI (Terminal Interface)**
Keyboard-driven interface for remote work and SSH sessions.
```bash
prism tui
```

### 💻 **CLI (Command Line)**
Scriptable interface for automation and power users.
```bash
prism workspace launch python-ml my-project --size L
```

---

## Essential Commands

### Template Management
```bash
prism templates                                    # List available environments
prism templates info python-ml                     # Get template details
prism workspace launch python-ml my-project        # Launch environment
```

### Instance Management
```bash
prism workspace list                               # Show running instances
prism workspace connect my-project                 # Get connection info
prism workspace stop my-project                    # Stop when not in use
prism workspace resume my-project                  # Resume later
prism workspace delete my-project                  # Remove completely
```

### Cost Optimization
```bash
prism workspace hibernate my-project               # Preserve work, reduce costs
prism workspace resume my-project                  # Resume hibernated instance
```

Auto-hibernation is configured per template (see each template's idle detection settings).

---

## Common Research Workflows

### Data Science Project
```bash
# Launch Jupyter environment
prism workspace launch python-ml data-analysis --size L

# Create shared storage
prism volume create shared-datasets

# Connect and start working
prism workspace connect data-analysis
# Opens: ssh user@ip-address -L 8888:localhost:8888
# Jupyter: http://localhost:8888
```

### R Statistical Analysis
```bash
# Launch R + RStudio environment
prism workspace launch r-rstudio-server stats-project

# Get RStudio connection
prism workspace connect stats-project
# Opens: http://ip-address:8787 (RStudio Server)
```

### Custom Environment
```bash
# Start with base template
prism workspace launch basic-ubuntu my-custom

# Customize your setup
prism workspace connect my-custom
# Install packages, configure tools

# Save for reuse as an AMI
prism ami save my-custom "my-custom-template"
```

---

## Troubleshooting

### "Daemon not running"
```bash
# Check daemon status (daemon usually auto-starts)
prism admin daemon status

# Restart daemon if needed
prism admin daemon stop
# Next command will auto-start a fresh daemon
prism templates
```

### "AWS credentials not found"
```bash
# Verify AWS configuration
aws sts get-caller-identity

# Reconfigure if needed
aws configure
```

### "Permission denied" errors
Make sure your AWS user has the required permissions. See our [AWS IAM Permissions](../admin-guides/AWS_IAM_PERMISSIONS.md) for complete IAM policies, or run:

```bash
./scripts/setup-iam-permissions.sh
```

### Instance launch fails
```bash
# Check AWS region and availability
aws ec2 describe-availability-zones

# Try different region
prism workspace launch python-ml my-project --region us-east-1
```

---

## Next Steps

- **Browse Templates**: Explore research environments with `prism templates`
- **Join Community**: Share templates and get help
- **Read Guides**: Detailed documentation in `/docs` folder
- **Cost Optimization**: Learn about hibernation and spot instances
- **Team Collaboration**: Set up shared storage and project management

**Need Help?** Open an issue on [GitHub](https://github.com/scttfrdmn/prism/issues) or check our documentation.

---

## Advanced Features

### Project Management
```bash
# Create research project
prism project create brain-study --budget 500

# Launch in project context
prism workspace launch neuroimaging analysis --project brain-study
```

### Custom AMIs
```bash
# Build an AMI from a template (pre-baked, fast launch)
prism ami create python-ml

# Save a running instance as an AMI
prism ami save my-project "My Custom Environment"

# Launch from your saved AMI
prism workspace launch --ami "My Custom Environment" fast-start
```

**🎯 Key Principle**: Prism defaults to success. Most commands work without options, with smart defaults for research computing.
