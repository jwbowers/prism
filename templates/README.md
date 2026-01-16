# Prism Templates

This directory contains Prism's template library for launching pre-configured cloud research environments.

## 📁 Directory Structure (as of v0.7.0)

```
templates/
├── base/           # Minimal base OS templates (maintained by Prism core team)
├── community/      # Community-contributed templates
├── deprecated/     # Old templates (will be removed in July 2026)
├── testing/        # Test templates for development
├── examples/       # Example templates for reference
└── README.md       # This file
```

## 🎯 Philosophy (Issue #429 Refactor)

As of January 2026, Prism's built-in template library has been **simplified**:

### ✅ What Prism Ships With
**Base OS templates only** (`templates/base/`):
- Amazon Linux 2023 (x86_64, arm64)
- Ubuntu 22.04 LTS (x86_64, arm64)
- Ubuntu 24.04 LTS (x86_64, arm64)

These are **minimal** - just OS + package manager + system user.

### ✅ Where Application Templates Live
1. **Community Templates** (`templates/community/`) - Contributed by users
2. **User Templates** (`~/.prism/templates/`) - Your custom templates
3. **Institutional Templates** - Hosted by your organization

## 🚀 Quick Start

### Launch a Minimal Instance
```bash
# Launch a minimal Ubuntu 24.04 instance
prism launch ubuntu-24-04-x86 my-instance

# Connect and install what you need
prism connect my-instance
sudo apt update && sudo apt install python3
```

### Use a Community Template
```bash
# List community templates
prism templates

# Launch R research environment
prism launch r-research-full-stack my-r-env
```

### Create Your Own Template
```yaml
# ~/.prism/templates/my-python-env.yml
name: "My Python Environment"
description: "Custom Python data science setup"
inherits: ["Ubuntu 24.04 LTS (x86_64)"]

packages:
  apt:
    - "python3"
    - "python3-pip"
  pip:
    - "numpy"
    - "pandas"
    - "matplotlib"

ports:
  - 22
  - 8888  # Jupyter
```

Then launch it:
```bash
prism launch my-python-env my-instance
```

## 📚 Documentation

### For Users
- **Using Templates**: `docs/user-guides/TEMPLATE_FORMAT.md`
- **Base Templates**: `templates/base/README.md`
- **Multi-User Setup**: `docs/user-guides/MULTI_USER_INSTANCE_SETUP.md` (Issue #432)
- **Custom AMIs**: `docs/user-guides/CUSTOM_AMI_WORKFLOW.md` (Issue #433)

### For Contributors
- **Community Guide**: `docs/development/COMMUNITY_TEMPLATE_GUIDE.md` (Issue #434)
- **Template Inheritance**: `docs/user-guides/TEMPLATE_FORMAT.md`
- **Testing Templates**: `docs/TESTING.md`

## 🔨 Template Inheritance

The recommended approach for creating templates:

```yaml
name: "R Research Environment"
inherits: ["Ubuntu 24.04 LTS (x86_64)"]

# Inherits: OS, package manager, ubuntu user, SSH access
# You add: Your specific packages and configuration

packages:
  apt:
    - "r-base"
    - "r-base-dev"
```

**Benefits**:
- ✅ Cleaner templates (focus on your additions)
- ✅ Automatic updates when base OS changes
- ✅ Composition over duplication
- ✅ Easier to maintain

## 📂 Subdirectories

### `base/` - Base Operating Systems
Minimal OS templates maintained by Prism core team.
- See `base/README.md` for details
- These are the foundation for all other templates
- Only essential system packages included

### `community/` - Community Templates
Application-specific templates contributed by the community.
- R research environments
- Python ML workstations
- Bioinformatics pipelines
- And more!

To contribute: See `docs/development/COMMUNITY_TEMPLATE_GUIDE.md`

### `deprecated/` - Old Templates
Previous built-in application templates.
- See `deprecated/README.md` for migration guide
- Will be removed in July 2026
- Available in git history after removal

### `testing/` - Test Templates
Templates used for automated testing and development.
- Not intended for production use
- May have test-specific configurations
- Used in CI/CD pipelines

### `examples/` - Example Templates
Reference templates showing various features.
- Template inheritance examples
- Package manager examples
- Advanced configuration examples

## 🎨 Template Format

Templates are YAML files with the following structure:

```yaml
name: "Template Display Name"
slug: "template-slug"
description: "Brief description of what this provides"
base: "ubuntu-24.04"  # or amazonlinux-2023, ubuntu-22.04

# Optional: Inherit from other templates
inherits: ["Ubuntu 24.04 LTS (x86_64)"]

# Connection type
connection_type: "ssh"  # or "desktop" for DCV, "web" for web apps

# Package manager (apt, dnf, conda, spack, or ami)
package_manager: "apt"

# Packages to install
packages:
  system:  # System packages (apt/dnf)
    - "git"
    - "wget"
  conda:   # Conda packages
    - "numpy"
    - "pandas"
  pip:     # Python packages
    - "requests"
  spack:   # Spack packages
    - "openmpi"

# Users to create
users:
  - name: "researcher"
    groups: ["sudo"]
    shell: "/bin/bash"

# Services to configure
services:
  - name: "jupyter"
    port: 8888
    start_on_boot: true

# Ports to open in security group
ports:
  - 22    # SSH
  - 8888  # Jupyter

# Instance defaults
instance_defaults:
  ports: [22]
  root_volume_gb: 20

version: "1.0"
tags:
  type: "application"
  purpose: "research"
```

## 🔍 Finding Templates

```bash
# List all available templates
prism templates

# Get detailed info about a template
prism templates info "template-name"

# Validate a template
prism templates validate path/to/template.yml

# Search by category
prism templates | grep "Machine Learning"
```

## ⚙️ Advanced Features

### Template Parameters
Some templates accept parameters at launch time:
```bash
prism launch python-ml my-instance \
  --param python_version=3.11 \
  --param install_pytorch=true
```

### Package Manager Override
Change the package manager at launch:
```bash
prism launch my-template my-instance --package-manager conda
```

### Size Scaling
Specify instance size:
```bash
prism launch my-template my-instance --size L
# Sizes: XS, S, M (default), L, XL
```

### Custom AMIs
Create reusable AMIs from configured instances:
```bash
# Configure your instance
prism launch ubuntu-24-04-x86 my-instance
prism connect my-instance
# ... install and configure ...

# Save as AMI
prism ami create my-instance --name "My Custom AMI"

# Launch from AMI (30 seconds vs 3+ minutes!)
prism launch --ami ami-abc123 quick-instance
```

See `docs/user-guides/CUSTOM_AMI_WORKFLOW.md` for details.

## 🐛 Troubleshooting

### Template Not Found
```bash
# Verify template exists
prism templates | grep "template-name"

# Check all template directories
echo $PRISM_TEMPLATE_DIR
ls ~/.prism/templates/
```

### Template Validation Errors
```bash
# Validate template syntax
prism templates validate template-name

# Common issues:
# - Missing required fields (name, base, packages)
# - Invalid YAML syntax
# - Invalid package manager
# - Circular inheritance
```

### Launch Failures
```bash
# Check template configuration
prism templates info template-name

# Verify AMI availability in your region
# Base templates should work in all regions

# Check CloudWatch logs for detailed errors
```

## 📊 Migration from v0.6.x

If you were using built-in application templates:

### Option 1: Use Community Templates
```bash
# Old (v0.6.x)
prism launch python-ml my-instance

# New (v0.7.0+)
prism launch python-ml-configurable my-instance
# (Community template in templates/community/)
```

### Option 2: Create Custom Template
```bash
# Copy deprecated template
cp templates/deprecated/python-ml-workstation.yml ~/.prism/templates/my-python-ml.yml

# Edit and customize
vim ~/.prism/templates/my-python-ml.yml

# Launch
prism launch my-python-ml my-instance
```

### Option 3: Manual Setup
```bash
# Launch minimal base
prism launch ubuntu-24-04-x86 my-instance

# Connect and install
prism connect my-instance
sudo apt install python3 python3-pip
pip3 install numpy pandas jupyter
```

See `templates/deprecated/README.md` for detailed migration guide.

## 🤝 Contributing

We welcome community template contributions!

1. **Fork** the repository
2. **Create** your template in `templates/community/`
3. **Test** thoroughly (`prism templates validate`)
4. **Document** usage and prerequisites
5. **Submit** a pull request

See `docs/development/COMMUNITY_TEMPLATE_GUIDE.md` for detailed guidelines.

## 📝 Related Issues & Documentation

- **#429**: Template library cleanup (this refactor)
- **#430**: R Research Full Stack template
- **#431**: R Research Stacked templates (experimental)
- **#432**: Multi-user instance setup guide
- **#433**: Custom AMI workflow guide
- **#434**: Community template contribution guide
- **#435**: Non-technical collaborator persona test

## 🔗 External Resources

- [Prism Documentation](https://github.com/scttfrdmn/prism)
- [Template Format Specification](docs/user-guides/TEMPLATE_FORMAT.md)
- [Community Templates Marketplace](https://github.com/scttfrdmn/prism/tree/main/templates/community)
- [Issue Tracker](https://github.com/scttfrdmn/prism/issues)
