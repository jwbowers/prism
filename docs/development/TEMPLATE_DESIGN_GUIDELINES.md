# Template Design Guidelines

**Version**: 1.0.0
**Date**: January 16, 2026
**Status**: Official Guidelines

## Overview

This document provides guidelines for creating and maintaining Prism templates. Following these guidelines ensures consistency, maintainability, and a better user experience.

## Quick Reference

| Template Tier | Purpose | Inherits From | Complexity | Example |
|---------------|---------|---------------|------------|---------|
| **Tier 1: Base OS** | Minimal OS + package manager | None | Simple | `ubuntu-24.04-x86.yml` |
| **Tier 2: Package Manager** | Add ecosystem tools | Tier 1 | Simple | `ubuntu-24.04-conda-x86.yml` |
| **Tier 3: Language** | Language runtime + basics | Tier 2 | Simple-Intermediate | `python-base-conda.yml` |
| **Tier 4: Domain** | Complete workstations | Tier 3 | Intermediate-Complex | `ml-workstation-gpu.yml` |

## When to Create a New Template

### ✅ Create a New Template When:

1. **Serving a distinct use case** not covered by existing templates
   - Example: "Bioinformatics Workstation" for genomics research
   - Example: "Quantum Chemistry Environment" for computational chemistry

2. **Providing a foundation for a template family**
   - Example: "Ubuntu 24.04 + Conda" as foundation for Python/ML stacks
   - Example: "R Base" as foundation for R ecosystem

3. **Offering significantly different tool combinations**
   - Example: GPU-optimized ML vs CPU-only ML workstation
   - Example: Jupyter notebooks vs RStudio IDE

### ❌ Don't Create a New Template When:

1. **Minor package differences** - use existing template and customize via launch options
   - Bad: Creating "Python ML + TensorFlow 2.14" vs "Python ML + TensorFlow 2.15"
   - Good: One "Python ML" template, users install specific versions as needed

2. **Personal preferences** - avoid templates for individual user customizations
   - Bad: "John's Data Science Setup"
   - Good: Contribute improvements to existing "Data Science Workstation"

3. **Temporary experiments** - use `templates/testing/` for experiments
   - Test templates not meant for production users

## Template Hierarchy Decision Tree

```
┌─────────────────────────────────────────────────┐
│ What are you creating?                          │
└─────────────────────────────────────────────────┘
                       │
        ┌──────────────┼──────────────┐
        │              │              │
    Base OS     Package Manager   Language/Domain
        │              │              │
        v              v              v
   Tier 1: OS     Tier 2: PM    Tier 3-4: App
   No inherits    Inherits OS   Inherits PM/Lang
   Simple         Simple         Complex
```

### Tier 1: Base Operating System Templates

**Purpose**: Provide minimal OS installation with package manager

**When to create**:
- Adding support for new OS (e.g., Rocky Linux, Debian)
- New OS version (e.g., Ubuntu 26.04 LTS)
- New architecture (e.g., RISC-V support)

**Template characteristics**:
```yaml
name: "{OS} {Version} LTS ({architecture})"
slug: "{os}-{version}-{arch}"
base: "{os}-{version}"              # Official base image
complexity: "simple"
package_manager: "apt|dnf|zypper"   # System package manager
inherits: []                        # No inheritance (foundation)

packages:
  system:
    - build-essential               # Minimal system packages only
    - curl
    - wget
    - git
```

**Examples**:
- `ubuntu-24.04-x86.yml`
- `amazon-linux-2023-arm.yml`
- `debian-12-x86.yml` (if created)

### Tier 2: Package Manager Foundation Templates

**Purpose**: Add package management ecosystem (Conda, Spack, Nix)

**When to create**:
- Foundation for language-specific stacks (Python/R via Conda)
- Scientific computing foundation (HPC via Spack)
- Reproducible environments (Nix)

**Template characteristics**:
```yaml
name: "{OS} {Version} + {PackageManager} ({architecture})"
slug: "{os}-{version}-{pm}-{arch}"
base: "{os}-{version}"
complexity: "simple"
package_manager: "conda|spack|nix"
inherits: ["{OS} {Version} LTS ({architecture})"]

post_install: |
  # Install package manager
  # Configure environment
  # Set up user paths
```

**Examples**:
- `ubuntu-24.04-conda-x86.yml`
- `ubuntu-24.04-spack-x86.yml` (if created)
- `rocky-9-conda-arm.yml` (if created)

### Tier 3: Language Foundation Templates

**Purpose**: Provide language runtime + essential packages

**When to create**:
- Base layer for language ecosystem (Python, R, Julia, Go)
- Language + key tool (Python + Jupyter, R + RStudio)
- Multi-stage language stacks (R Base → R + RStudio → R Publishing)

**Template characteristics**:
```yaml
name: "{Language} {Component}"
slug: "{language}-{component}"
base: "{os}-{version}"
complexity: "simple" | "intermediate"
package_manager: "apt|conda"  # Match parent
inherits: ["{OS} + {PackageManager}"]  # Or another language template

# Add language runtime
# Add essential packages only
# Keep focused on language foundation
```

**Examples**:
- `r-base-ubuntu24.yml` - R foundation
- `r-rstudio-server.yml` - R + IDE
- `python-base-conda.yml` - Python foundation (to create)
- `python-jupyter.yml` - Python + notebooks (to create)

### Tier 4: Domain-Specific Workstation Templates

**Purpose**: Complete research environment for specific domain

**When to create**:
- Comprehensive domain solution (bioinformatics, geospatial, ML)
- Multi-language integrated environments
- Specialized hardware requirements (GPU, high-memory)

**Template characteristics**:
```yaml
name: "{Domain} Workstation" | "{Purpose} Environment"
slug: "{domain}-workstation"
base: "{os}-{version}"
complexity: "intermediate" | "complex"
package_manager: # Match parent
inherits: ["{Language} {Component}"]  # Or multiple languages

# Domain-specific tools
# Integrated workflow
# Rich user documentation
```

**Examples**:
- `r-publishing-stack.yml` - R + Quarto + LaTeX + Python
- `bioinformatics-workstation.yml` (to create)
- `ml-workstation-gpu.yml` (to create)

## Naming Conventions

### Template File Names

**Format**: `{descriptive-name}.yml`
- Use lowercase
- Use hyphens (not underscores)
- Be descriptive but concise
- Include key differentiators

**Examples**:
- ✅ `ubuntu-24.04-x86.yml` - Clear OS, version, architecture
- ✅ `python-ml-workstation-gpu.yml` - Clear purpose and variant
- ❌ `my-template.yml` - Not descriptive
- ❌ `python_ml.yml` - Use hyphens, not underscores

### Template Display Names (`name` field)

**Tier 1 (Base OS)**:
- Format: `"{OS} {Version} LTS ({architecture})"`
- Example: `"Ubuntu 24.04 LTS (x86_64)"`

**Tier 2 (Package Manager)**:
- Format: `"{OS} {Version} + {PackageManager} ({architecture})"`
- Example: `"Ubuntu 24.04 + Conda (x86_64)"`

**Tier 3 (Language)**:
- Format: `"{Language} {Component}"`
- Example: `"R + RStudio Server"`
- Example: `"Python + Jupyter Lab"`

**Tier 4 (Domain)**:
- Format: `"{Domain} Workstation"` or `"{Purpose} Environment"`
- Example: `"Bioinformatics Workstation"`
- Example: `"ML Workstation (GPU)"`

### Template Slugs (`slug` field)

**Format**: Lowercase, hyphenated, no special characters
- Tier 1: `"{os}-{version}-{arch}"`
- Tier 2: `"{os}-{version}-{pm}-{arch}"`
- Tier 3: `"{language}-{component}"`
- Tier 4: `"{domain}-workstation"` or `"{purpose}-env"`

## Template Structure Best Practices

### Required Fields

```yaml
# === Identity ===
name: "Template Display Name"
slug: "template-slug"
description: "One-line summary (< 80 chars)"

# === Technical ===
base: "ubuntu-24.04"              # Base OS
author: "Prism Team" | "Your Name"
version: "1.0.0"                  # Semantic versioning
architecture: "x86_64" | "arm64"  # Target architecture

# === Classification ===
connection_type: "ssh" | "desktop"
complexity: "simple" | "intermediate" | "complex"
category: "Research Tools" | "Machine Learning" | etc.
domain: "research" | "ml" | "hpc" | etc.

# === Visual ===
icon: "📊"                        # Emoji icon
color: "#276DC3"                  # Hex color
popular: true | false             # Show in popular section

# === Infrastructure ===
package_manager: "apt" | "dnf" | "conda"
inherits: ["Parent Template Name"]  # Optional but recommended

# === Resources ===
users:
  - name: "primary-user"
    groups: ["sudo"]
    shell: "/bin/bash"

services: []                      # Optional
ports:
  - 22                            # SSH always included

# === Installation ===
post_install: |
  #!/bin/bash
  set -e
  # Installation script

# === Operational ===
idle_detection:
  enabled: true
  idle_threshold_minutes: 30
  check_processes: ["process-name"]

instance_defaults:
  ports: [22]
  root_volume_gb: 20

# === Metadata ===
components:
  tool_version: "1.2.3"           # Track component versions

tags:
  type: "research"
  purpose: "data-science"
  level: "beginner" | "intermediate" | "advanced"
```

### Field Guidelines

**`description`**:
- One line, < 80 characters
- Clear, concise summary
- Include key differentiator
- Examples:
  - ✅ "Minimal R environment with essential packages"
  - ✅ "Complete ML workstation with GPU support (CUDA 12.2)"
  - ❌ "A template for doing data science stuff"

**`long_description`**:
- Multi-line, detailed explanation
- Include what's installed
- Mention inheritance (if applicable)
- List use cases
- Example:
  ```yaml
  long_description: |
    Full-featured R research environment built via template stacking.

    Inherits from R + RStudio Server, adding:
    - Quarto 1.6.33 for scientific publishing
    - TeX Live 2024 (full LaTeX distribution)
    - Python 3.12 with Jupyter Lab

    Perfect for:
    - Publishing research papers
    - Mixed R/Python workflows
    - Collaborative analysis
  ```

**`inherits`**:
- Use display names (from parent's `name` field)
- Single parent (linear inheritance)
- Clear chain: Tier 1 → Tier 2 → Tier 3 → Tier 4
- Example:
  ```yaml
  # Tier 3 inherits from Tier 2
  inherits: ["Ubuntu 24.04 + Conda (x86_64)"]

  # Tier 4 inherits from Tier 3
  inherits: ["Python + Jupyter Lab"]
  ```

**`post_install`**:
- Always start with `#!/bin/bash` and `set -e`
- Use clear section comments
- Echo progress messages
- Verify installations
- Clean up temporary files
- Example:
  ```bash
  post_install: |
    #!/bin/bash
    set -e

    echo "🔧 Installing Component X..."

    # Install packages
    apt-get update
    apt-get install -y package-name

    # Verify installation
    which package-name || exit 1
    echo "✅ Component X installed successfully"

    # Cleanup
    apt-get clean
  ```

**`packages`**:
- Group by purpose (use YAML objects)
- Clear comments for non-obvious packages
- Example:
  ```yaml
  packages:
    # Core development tools
    dev_tools:
      - build-essential
      - cmake
      - git

    # Language-specific
    python:
      - python3.12
      - python3-pip

    # Domain-specific
    bioinformatics:
      - samtools       # SAM/BAM processing
      - bcftools       # VCF/BCF processing
  ```

## Inheritance Best Practices

### Do's ✅

1. **Linear inheritance chains**: Tier 1 → Tier 2 → Tier 3 → Tier 4
   ```yaml
   Ubuntu 24.04 (x86_64)
     └─> Ubuntu 24.04 + Conda (x86_64)
           └─> Python + Jupyter
                 └─> ML Workstation (GPU)
   ```

2. **Reuse foundations**: Don't duplicate base layers
   ```yaml
   # Good: Reuse Conda foundation
   name: "Python ML Workstation"
   inherits: ["Ubuntu 24.04 + Conda (x86_64)"]
   ```

3. **Document inheritance**: Explain what each layer adds
   ```yaml
   long_description: |
     Inherits from Python + Jupyter, adding:
     - PyTorch 2.1 (GPU support)
     - TensorFlow 2.14
     - scikit-learn, pandas, matplotlib
   ```

4. **Test inheritance chain**: Validate each layer independently
   ```bash
   # Test Layer 1
   prism workspace launch "Ubuntu 24.04 (x86_64)" test-tier1

   # Test Layer 2
   prism workspace launch "Ubuntu 24.04 + Conda (x86_64)" test-tier2

   # Test Layer 3
   prism workspace launch "Python + Jupyter" test-tier3
   ```

### Don'ts ❌

1. **Diamond inheritance**: Multiple parents not supported
   ```yaml
   # ❌ Not supported
   inherits: ["Python Base", "R Base"]
   ```

2. **Circular dependencies**: Template cannot inherit from child
   ```yaml
   # ❌ Circular dependency
   # Template A inherits: [Template B]
   # Template B inherits: [Template A]
   ```

3. **Deep nesting**: Keep chains < 5 levels
   ```yaml
   # ❌ Too deep (6 levels)
   OS → PM → Lang → IDE → Domain → Specialty → SubSpecialty

   # ✅ Good depth (4 levels)
   OS → PM → Lang → Domain
   ```

4. **Redefining inherited fields**: Don't duplicate users, ports
   ```yaml
   # Parent defines:
   users:
     - name: researcher

   # ❌ Child should NOT redefine
   users:
     - name: researcher  # Causes "duplicate user" error

   # ✅ Child should omit or comment
   # users inherited from parent
   ```

## Package Manager Selection

### When to use APT (System Packages)

**Use APT when**:
- System-level tools (compilers, libraries)
- OS integration required (systemd services)
- Native performance critical
- Template targets single OS

**Examples**:
- System libraries: `libssl-dev`, `libcurl4-openssl-dev`
- Compilers: `gcc`, `g++`, `gfortran`
- Services: `postgresql`, `nginx`

### When to use Conda (Python/R Ecosystem)

**Use Conda when**:
- Python or R focused environment
- Cross-platform reproducibility important
- User-level package management preferred
- Scientific computing packages

**Examples**:
- Python packages: `pytorch`, `tensorflow`, `pandas`
- R packages: `r-tidyverse`, `r-ggplot2`
- Scientific libraries: `numpy`, `scipy`, `matplotlib`

### When to use Spack (HPC/Scientific)

**Use Spack when**:
- High-performance computing environment
- Multiple compiler/library versions needed
- Optimized builds for specific hardware

**Examples**:
- MPI implementations: `openmpi`, `mpich`
- Scientific libraries: `petsc`, `trilinos`
- Compilers: Multiple gcc/clang versions

## Testing Requirements

### Validation Tests (Required)

All templates must pass:

```bash
# 1. Syntax validation
prism templates validate

# 2. Template info check
prism templates info "Template Name"

# 3. Dry-run launch
prism workspace launch "Template Name" test-instance --dry-run
```

### Integration Tests (Recommended)

For complex templates (Tier 3-4):

```go
func TestTemplate_YourTemplateName(t *testing.T) {
    ctx := NewTestContext(t)
    registry := fixtures.NewFixtureRegistry(t, ctx.Client)

    // Launch instance
    instance, err := fixtures.CreateTestInstance(t, registry,
        fixtures.CreateTestInstanceOptions{
            Template: "Your Template Name",
            Name:     "test-instance",
            Size:     "M",
        })
    AssertNoError(t, err, "Template should launch successfully")

    // Test key functionality
    // (e.g., service accessible, tools installed)
}
```

### Manual Testing Checklist

Before submitting template:

- [ ] Launch instance successfully
- [ ] SSH connection works
- [ ] Key services start automatically
- [ ] Tools/packages installed correctly
- [ ] Post-install script completes without errors
- [ ] User can perform intended workflow
- [ ] Documentation accurate and complete
- [ ] Cost estimate reasonable for use case

## Documentation Requirements

### Template Documentation

Each Tier 3-4 template should have:

1. **Inline documentation** (in YAML):
   - `description`: One-line summary
   - `long_description`: Detailed explanation
   - Comments in `packages` and `post_install`

2. **User guide** (optional for complex templates):
   - `docs/user-guides/{TEMPLATE_NAME}_GUIDE.md`
   - Quick start (5 min)
   - What's included
   - Usage examples
   - Troubleshooting

3. **README section** (for template families):
   - Update `docs/user-guides/TEMPLATES_OVERVIEW.md`
   - Add template to appropriate category
   - Link to detailed guide

### Example Documentation Structure

```markdown
# {Template Name} Guide

## Overview
Brief description of template purpose and target users.

## Quick Start (⏱️ 5 minutes)
Step-by-step launch instructions.

## What's Included
List of installed tools, packages, services.

## Usage Examples
2-3 concrete examples of common workflows.

## Troubleshooting
Common issues and solutions.

## Related Templates
Links to similar or alternative templates.
```

## Version Management

### Semantic Versioning

Templates use semantic versioning:

```yaml
version: "MAJOR.MINOR.PATCH"
```

- **MAJOR**: Breaking changes (incompatible with previous version)
- **MINOR**: New features (backwards compatible)
- **PATCH**: Bug fixes, documentation updates

### Component Versions

Track installed component versions:

```yaml
components:
  r_version: "4.4.2"
  rstudio_server_version: "2024.12.0"
  quarto_version: "1.6.33"
  python_version: "3.12"
  cuda_version: "12.2"  # If applicable
```

### Deprecation Process

When deprecating templates:

1. **Add deprecation notice** to description
2. **Point to replacement** template
3. **Keep template** for 3 months minimum
4. **Move to** `templates/deprecated/` after grace period
5. **Update documentation** with migration guide

## Common Patterns

### Pattern: Monolithic vs Stacked

**Monolithic** (all-in-one):
```yaml
name: "Complete Research Environment"
inherits: ["Ubuntu 24.04 (x86_64)"]
# Everything in one template
```

**Stacked** (layered):
```yaml
# Layer 1
name: "R Base"
inherits: ["Ubuntu 24.04 (x86_64)"]

# Layer 2
name: "R + RStudio"
inherits: ["R Base"]

# Layer 3
name: "R Publishing Stack"
inherits: ["R + RStudio"]
```

**When to use each**:
- Monolithic: Users always need everything, simplicity > flexibility
- Stacked: Users need choices, maintainability important

### Pattern: Multi-User Setup

```yaml
users:
  - name: "researcher"    # Primary user
    groups: ["sudo"]
    shell: "/bin/bash"

  # Additional users created in post_install
  # Document in long_description

post_install: |
  # Create shared directory
  mkdir -p /home/shared/projects
  chmod 2775 /home/shared/projects
  chgrp sudo /home/shared/projects
```

### Pattern: Web Service Template

```yaml
services:
  - name: "jupyter"
    port: 8888
    type: "web"
    description: "Jupyter Notebook Server"
    start_command: "jupyter lab --ip=0.0.0.0 --no-browser"
    check_command: "pgrep -f jupyter-lab"

ports:
  - 22    # SSH
  - 8888  # Jupyter

long_description: |
  Access Jupyter Lab at: http://YOUR_IP:8888
  Default token shown in launch output.
```

## Checklist for New Templates

Before submitting a new template:

### Technical
- [ ] Template validates successfully (`prism templates validate`)
- [ ] Inherits from appropriate parent template
- [ ] `post_install` script tested and works
- [ ] All required fields present
- [ ] Naming follows conventions
- [ ] Package manager consistent with parent
- [ ] Services start automatically
- [ ] Idle detection configured appropriately

### Documentation
- [ ] Clear, concise `description`
- [ ] Detailed `long_description`
- [ ] Use cases documented
- [ ] Inheritance chain explained (if applicable)
- [ ] Cost estimate provided (in description or guide)
- [ ] User guide created (for complex templates)

### Testing
- [ ] Manual test: template launches successfully
- [ ] Manual test: key functionality works
- [ ] Manual test: user workflow validated
- [ ] Integration test created (for complex templates)
- [ ] Edge cases tested (small/large instances, spot, etc.)

### Compliance
- [ ] Follows template hierarchy (Tier 1-4)
- [ ] Consistent with existing template family
- [ ] No unnecessary duplication
- [ ] Reasonable resource requirements
- [ ] Secure by default (no hardcoded passwords)

## Getting Help

- **Template issues**: Open issue with `templates` label
- **Design questions**: Consult `TEMPLATE_REFACTORING_PLAN.md`
- **Examples**: Reference existing templates in `templates/`
- **Community**: Discussions in GitHub or Slack

---

**Document Version**: 1.0.0
**Last Updated**: January 16, 2026
**Related**: TEMPLATE_REFACTORING_PLAN.md, COMMUNITY_TEMPLATE_GUIDE.md
