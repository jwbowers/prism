# Base OS Templates

This directory contains Prism's **minimal base operating system templates**. These are the foundation upon which all other templates should build.

## Available Base Templates

### Amazon Linux 2023
- **amazon-linux-2023-x86.yml** - x86_64 architecture
- **amazon-linux-2023-arm.yml** - arm64 architecture (Graviton)

**Features**:
- DNF package manager
- Minimal system packages (git, wget, curl, vim, htop)
- Default user: `ec2-user` (wheel group)
- 20GB root volume

### Ubuntu 22.04 LTS
- **ubuntu-22.04-x86.yml** - x86_64 architecture
- **ubuntu-22.04-arm.yml** - arm64 architecture

**Features**:
- APT package manager
- Minimal system packages + build-essential
- Default user: `ubuntu` (sudo group)
- 20GB root volume
- Long-term support until 2027

### Ubuntu 24.04 LTS
- **ubuntu-24.04-x86.yml** - x86_64 architecture
- **ubuntu-24.04-arm.yml** - arm64 architecture

**Features**:
- APT package manager
- Minimal system packages + build-essential
- Default user: `ubuntu` (sudo group)
- 20GB root volume
- Long-term support until 2029

## Design Philosophy

Base templates follow these principles:

### 1. Minimal by Default
- Only include essential system packages
- No application software (Python, R, databases, etc.)
- No development frameworks
- No GUI/desktop environments

### 2. OS + Package Manager Only
Each base template provides:
- Operating system configuration
- Package manager (APT, DNF)
- Default system user
- SSH access (port 22)
- Basic utilities (git, wget, curl, vim, htop)

### 3. Architecture-Specific
- Separate templates for x86_64 and arm64
- Allows architecture-specific optimizations
- Clear naming convention

### 4. Foundation for Inheritance
Base templates are designed to be inherited by application templates:

```yaml
name: "My Application Environment"
inherits: ["Ubuntu 24.04 LTS (x86_64)"]

packages:
  apt:
    - "postgresql-client"
    - "python3-pip"
  pip:
    - "django"
    - "celery"
```

## Usage

### Direct Launch (Minimal Instance)
```bash
# Launch a minimal Ubuntu instance
prism launch ubuntu-24-04-x86 my-minimal-instance

# Connect and install what you need
prism connect my-minimal-instance
sudo apt update
sudo apt install python3 python3-pip
```

### As a Base for Custom Templates
```yaml
# ~/.prism/templates/my-template.yml
name: "My Custom Environment"
inherits: ["Ubuntu 24.04 LTS (x86_64)"]

packages:
  apt:
    - "your-packages-here"
```

## Architecture Selection

### When to Use x86_64
- Need maximum software compatibility
- Using software that doesn't support ARM
- Working with legacy applications
- GPU workloads (most GPU instances are x86_64)

### When to Use arm64 (Graviton)
- Cost optimization (Graviton is ~20% cheaper)
- Better price/performance for general workloads
- Modern software that supports ARM
- Energy efficiency matters

## Which Base Template to Choose?

### Amazon Linux 2023
**Best for**:
- AWS-optimized workloads
- Integration with AWS services
- DNF package manager preference
- Enterprise environments

**Considerations**:
- Smaller package ecosystem than Ubuntu
- AWS-specific optimizations
- RPM-based package management

### Ubuntu 22.04 LTS
**Best for**:
- Stability and long-term support
- Wide software compatibility
- Familiar APT package manager
- Maximum package availability

**Considerations**:
- LTS until 2027
- Mature ecosystem
- Ubuntu-specific configurations

### Ubuntu 24.04 LTS
**Best for**:
- Latest features and packages
- Modern development
- Long-term projects (support until 2029)
- Most recent Ubuntu LTS

**Considerations**:
- Newest Ubuntu LTS
- Latest package versions
- Longest support timeline

## Creating Application Templates

See `docs/development/COMMUNITY_TEMPLATE_GUIDE.md` (Issue #434) for comprehensive guidance on creating templates that inherit from these base templates.

### Quick Example

```yaml
name: "Python Data Science"
description: "Python environment with data science tools"
inherits: ["Ubuntu 24.04 LTS (x86_64)"]

packages:
  apt:
    - "python3"
    - "python3-pip"
  pip:
    - "numpy"
    - "pandas"
    - "matplotlib"
    - "jupyter"

services:
  - name: "jupyter"
    port: 8888
    command: "jupyter notebook --ip=0.0.0.0 --no-browser"

ports:
  - 22
  - 8888
```

## Testing Base Templates

```bash
# Validate template syntax
prism templates validate templates/base/ubuntu-24.04-x86.yml

# Test launch
prism launch ubuntu-24-04-x86 test-instance

# Verify
prism connect test-instance
uname -a
lsb_release -a

# Cleanup
prism delete test-instance
```

## Contributing

Base templates are maintained by the Prism core team. If you believe a package should be added to or removed from a base template, please open an issue with:

- Clear justification
- Use case explanation
- Impact analysis

Remember: Base templates should remain **minimal**. Most additions belong in application templates, not base templates.
