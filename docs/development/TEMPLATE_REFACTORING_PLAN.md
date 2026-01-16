# Template Library Refactoring Plan

**Issue**: #408
**Date**: January 16, 2026
**Status**: Proposal

## Executive Summary

The current template library (40 templates) has grown organically and would benefit from systematic organization. This document proposes a refactoring plan to create a more foundational, consistent, and maintainable template structure.

## Current State Analysis

### Directory Structure

```
templates/
├── base/           # 6 templates - OS foundations (Ubuntu 22/24, Amazon Linux)
├── community/      # 9 templates - User-facing templates
├── deprecated/     # 17 templates - Old templates kept for reference
├── examples/       # 1 template - AMI-based example
└── testing/        # 6 templates - Test fixtures
```

### Base Templates (Well-Organized ✅)

```
amazon-linux-2023-arm.yml       # DNF, arm64
amazon-linux-2023-x86.yml       # DNF, x86_64
ubuntu-22.04-arm.yml            # APT, arm64
ubuntu-22.04-x86.yml            # APT, x86_64
ubuntu-24.04-arm.yml            # APT, arm64
ubuntu-24.04-x86.yml            # APT, x86_64
```

**Strengths**:
- Clear naming: OS + version + architecture
- Consistent structure
- Minimal packages (foundation only)
- Well-documented

### Community Templates (Needs Improvement ⚠️)

```
base-ubuntu-apt.yml                  # Ubuntu 22.04, APT, no inherits
python-ml-configurable.yml           # Ubuntu 22.04, Conda, no inherits
r-base-ubuntu24.yml                  # Ubuntu 24.04, APT, inherits Ubuntu 24.04 x86
r-rstudio-server.yml                 # Ubuntu 24.04, APT, inherits R Base
r-publishing-stack.yml               # Ubuntu 24.04, APT, inherits R + RStudio
r-research-full-stack.yml            # Ubuntu 24.04, APT, inherits Ubuntu 24.04 x86
ubuntu-desktop-explicit.yml          # Ubuntu 22.04, Conda, no inherits
ubuntu-python-r-ml.yml               # Ubuntu 22.04, Conda, no inherits
ultimate-research-workstation.yml    # Ubuntu 22.04, Conda, no inherits
```

**Issues Identified**:

1. **Ubuntu Version Inconsistency**: Mix of 22.04 and 24.04
   - R templates use 24.04 (newer, good choice)
   - Older templates use 22.04
   - Recommendation: Migrate to 24.04 (LTS, wider support)

2. **Package Manager Inconsistency**: Mix of APT and Conda
   - R templates use APT (system packages)
   - Python/ML templates use Conda (Python ecosystem)
   - Both are valid but not clearly documented

3. **Inheritance Patterns**: Inconsistent use of `inherits`
   - New R templates use inheritance (good pattern)
   - Older templates duplicate everything (monolithic)
   - Opportunity: Refactor older templates to use inheritance

4. **Naming Conventions**: Mix of technical and descriptive names
   - Technical: "base-ubuntu-apt", "python-ml-configurable"
   - Descriptive: "Ultimate Research Workstation", "Collaborative Research Workspace"
   - Inconsistent capitalization and hyphenation

5. **Missing Foundation Layers**: No clear package manager base templates
   - Could benefit from explicit "Ubuntu 24.04 + Conda" base template
   - Would enable Conda-based template families

## Proposed Template Organization

### Tier 1: Operating System Foundations (Existing - Keep As Is ✅)

```
templates/base/
├── amazon-linux-2023-arm.yml
├── amazon-linux-2023-x86.yml
├── ubuntu-22.04-arm.yml
├── ubuntu-22.04-x86.yml
├── ubuntu-24.04-arm.yml
└── ubuntu-24.04-x86.yml
```

**Purpose**: Minimal OS installation with package manager
**Characteristics**:
- No inherits (foundation layer)
- OS + package manager only
- Minimal system packages

### Tier 2: Package Manager Foundations (New Layer 🆕)

```
templates/foundations/
├── ubuntu-24.04-apt-x86.yml         # APT-based stack foundation
├── ubuntu-24.04-apt-arm.yml         # APT ARM variant
├── ubuntu-24.04-conda-x86.yml       # Conda-based stack foundation
└── ubuntu-24.04-conda-arm.yml       # Conda ARM variant
```

**Purpose**: Add package management ecosystem (Conda, Spack, etc.)
**Characteristics**:
- Inherits from Tier 1 (base OS)
- Adds package manager tools (Conda, Spack, etc.)
- Provides foundation for language-specific stacks

**Example**: `ubuntu-24.04-conda-x86.yml`
```yaml
name: "Ubuntu 24.04 + Conda (x86_64)"
slug: "ubuntu-24-04-conda-x86"
inherits: ["Ubuntu 24.04 LTS (x86_64)"]
base: "ubuntu-24.04"
package_manager: "conda"
complexity: "simple"

post_install: |
  # Install Miniconda
  wget https://repo.anaconda.com/miniconda/Miniconda3-latest-Linux-x86_64.sh
  bash Miniconda3-latest-Linux-x86_64.sh -b -p /opt/conda
  rm Miniconda3-latest-Linux-x86_64.sh

  # Configure conda
  /opt/conda/bin/conda init
  /opt/conda/bin/conda config --set auto_activate_base true
```

### Tier 3: Language Foundations (Refactored Community Templates)

```
templates/languages/
# R Family (Already Good ✅)
├── r-base-ubuntu24.yml              # R 4.4+ foundation
├── r-rstudio-server.yml             # + RStudio Server web IDE
└── r-publishing-stack.yml           # + Quarto/LaTeX/Python/tools

# Python Family (To Refactor)
├── python-base-conda.yml            # Python 3.12 + Conda foundation
├── python-jupyter.yml               # + Jupyter Lab
└── python-ml-stack.yml              # + ML libraries (PyTorch, TensorFlow)

# Multi-Language (To Create)
├── data-science-minimal.yml         # Python + R basics
└── data-science-full.yml            # Python + R + Julia
```

**Purpose**: Language runtime + essential packages
**Characteristics**:
- Inherits from Tier 2 (package manager foundation)
- Adds language runtime (R, Python, Julia)
- Minimal language packages (framework only)

### Tier 4: Domain-Specific Workstations (User-Facing)

```
templates/workstations/
# Research Computing
├── bioinformatics-workstation.yml   # Genomics, RNA-seq tools
├── computational-chemistry.yml      # Quantum chemistry, molecular dynamics
├── geospatial-workstation.yml       # QGIS, GDAL, PostGIS
└── high-performance-computing.yml   # MPI, compilers, schedulers

# Data Science & ML
├── ml-workstation-cpu.yml           # ML on CPU instances
├── ml-workstation-gpu.yml           # ML on GPU instances (CUDA)
└── statistical-analysis.yml         # R + Python + statistical tools

# Specialized
├── desktop-workstation.yml          # Full Ubuntu desktop (Nice DCV)
└── collaborative-research.yml       # Multi-user setup
```

**Purpose**: Complete research environments for specific domains
**Characteristics**:
- Inherits from Tier 3 (language stack)
- Adds domain-specific tools
- May combine multiple language stacks
- Rich user-facing documentation

## Naming Convention Guidelines

### Tier 1: Base OS Templates
**Format**: `{os}-{version}-{arch}.yml`
- Examples: `ubuntu-24.04-x86.yml`, `amazon-linux-2023-arm.yml`
- Name field: `"{OS} {Version} LTS ({arch})"`

### Tier 2: Package Manager Foundations
**Format**: `{os}-{version}-{package-manager}-{arch}.yml`
- Examples: `ubuntu-24.04-conda-x86.yml`, `ubuntu-24.04-apt-arm.yml`
- Name field: `"{OS} {Version} + {PackageManager} ({arch})"`

### Tier 3: Language Foundations
**Format**: `{language}-{component}-{variant}.yml`
- Examples: `r-base-ubuntu24.yml`, `python-jupyter.yml`
- Name field: `"{Language} {Component}"`
- Use descriptive component names: "Base", "+ RStudio Server", "+ Jupyter"

### Tier 4: Domain Workstations
**Format**: `{domain}-workstation.yml` or `{purpose}-{variant}.yml`
- Examples: `bioinformatics-workstation.yml`, `ml-workstation-gpu.yml`
- Name field: Full descriptive name: "Bioinformatics Workstation", "ML Workstation (GPU)"

## Migration Strategy

### Phase 1: Foundation Layers (Week 1) 🆕

**Create Tier 2 templates** (package manager foundations):
1. `ubuntu-24.04-conda-x86.yml` - Conda foundation for Python/ML stacks
2. `ubuntu-24.04-conda-arm.yml` - ARM variant
3. `ubuntu-24.04-apt-x86.yml` - Explicit APT foundation (optional, mostly redundant with base)
4. `ubuntu-24.04-apt-arm.yml` - ARM variant (optional)

**Action**: Create new templates in `templates/foundations/`

### Phase 2: Refactor Python/ML Templates (Week 2) ♻️

**Refactor existing templates to use inheritance**:

1. **python-ml-configurable.yml** → Refactor to use Conda foundation
   ```yaml
   name: "Python ML Environment"
   slug: "python-ml"
   inherits: ["Ubuntu 24.04 + Conda (x86_64)"]
   complexity: "intermediate"

   # Add ML packages via Conda
   packages:
     conda:
       - pytorch
       - tensorflow
       - scikit-learn
       - pandas
       - matplotlib
   ```

2. **ubuntu-python-r-ml.yml** → Refactor as multi-language stack
   ```yaml
   name: "Data Science Workstation (Python + R)"
   slug: "data-science-workstation"
   inherits: ["Python ML Environment"]
   complexity: "complex"

   # Add R via Conda
   packages:
     conda:
       - r-base
       - r-tidyverse
       - r-ggplot2
   ```

3. **ultimate-research-workstation.yml** → Simplify and clarify purpose
   - Current: 500+ lines, everything included
   - Proposed: Inherit from multiple foundations, add only unique tools
   - Better name: "Collaborative Research Workstation" (emphasizes multi-user)

**Action**: Move to `templates/languages/` or `templates/workstations/` as appropriate

### Phase 3: Documentation & Guidelines (Week 3) 📝

**Create comprehensive documentation**:

1. **Template Design Guidelines** (`docs/development/TEMPLATE_DESIGN_GUIDELINES.md`):
   - When to create new template vs use inheritance
   - How to choose base template
   - Naming conventions
   - Package manager selection criteria
   - Testing requirements

2. **Template Family Documentation** (`docs/user-guides/TEMPLATE_FAMILIES.md`):
   - Visual hierarchy diagrams
   - "Which template should I use?" decision tree
   - Comparison tables (monolithic vs stacked)

3. **Update existing docs**:
   - `docs/development/COMMUNITY_TEMPLATE_GUIDE.md` - Add inheritance patterns
   - `docs/user-guides/R_RESEARCH_TEMPLATE_GUIDE.md` - Update with new structure
   - `CLAUDE.md` - Update template inheritance section

**Action**: Create documentation, update existing guides

### Phase 4: Migration Plan Communication (Week 4) 📢

**Deprecation and Migration Path**:

1. **Mark old templates as deprecated** (keep for backwards compatibility):
   - Add deprecation notice to template descriptions
   - Point users to new equivalent templates
   - Keep old templates for 2-3 months

2. **Create migration guide** for users:
   - Old template → new template mapping
   - Launch command changes (if any)
   - Benefits of migrating (faster, better maintained)

3. **Update examples and tutorials**:
   - Use new template names in documentation
   - Update quick start guides
   - Revise persona walkthrough documents

**Action**: Communication plan, migration guide

## Implementation Checklist

### Week 1: Foundation Templates
- [ ] Create `templates/foundations/` directory
- [ ] Create `ubuntu-24.04-conda-x86.yml`
- [ ] Create `ubuntu-24.04-conda-arm.yml`
- [ ] Test Conda foundations work correctly
- [ ] Validate inheritance from base OS templates

### Week 2: Refactor Python/ML
- [ ] Refactor `python-ml-configurable.yml` to use Conda foundation
- [ ] Refactor `ubuntu-python-r-ml.yml` to use inheritance
- [ ] Simplify `ultimate-research-workstation.yml`
- [ ] Move refactored templates to appropriate tier directories
- [ ] Test all refactored templates launch successfully

### Week 3: Documentation
- [ ] Create `TEMPLATE_DESIGN_GUIDELINES.md`
- [ ] Create `TEMPLATE_FAMILIES.md`
- [ ] Update `COMMUNITY_TEMPLATE_GUIDE.md`
- [ ] Update `R_RESEARCH_TEMPLATE_GUIDE.md`
- [ ] Update `CLAUDE.md`
- [ ] Create visual hierarchy diagrams

### Week 4: Migration & Communication
- [ ] Add deprecation notices to old templates
- [ ] Create migration guide
- [ ] Update examples and tutorials
- [ ] Test backwards compatibility
- [ ] Announce changes to users

## Success Metrics

- ✅ Clear 4-tier template hierarchy
- ✅ Consistent naming conventions across all templates
- ✅ >80% of community templates use inheritance
- ✅ Comprehensive documentation for template authors
- ✅ Zero breaking changes for existing users
- ✅ Improved user experience: easier to find right template

## Benefits

**For Template Authors**:
- Clear patterns to follow
- Less code duplication (inheritance)
- Easier to maintain and update
- Better testing (test layers independently)

**For Users**:
- Clearer template hierarchy (foundational → specialized)
- Easier to choose correct template
- More flexible (choose layer depth)
- Better documentation

**For Maintenance**:
- Easier to update base OS versions
- Package manager updates propagate automatically
- Reduced duplication reduces bugs
- Clear ownership and responsibility per tier

## Timeline

- **Week 1** (Jan 20-26): Create foundation layer templates
- **Week 2** (Jan 27-Feb 2): Refactor Python/ML templates
- **Week 3** (Feb 3-9): Documentation and guidelines
- **Week 4** (Feb 10-16): Migration plan and communication

**Total Duration**: 4 weeks
**Impact**: Post-v0.6.0 quality improvement

## Related Work

- **Issue #431**: Stacked R templates (completed) - Proof of concept for inheritance
- **Issue #430**: Monolithic R template (completed) - Comparison baseline
- **Issue #429**: Template library refactoring docs (completed) - Background context
- **TEMPLATE_STACKING_ANALYSIS.md**: Analysis of approaches - Design validation

## Open Questions

1. **Should we create explicit APT foundation templates?**
   - Pro: Symmetry with Conda foundations
   - Con: Redundant with base OS (APT is system default)
   - **Recommendation**: No, base OS templates serve this purpose

2. **How to handle ARM architecture templates?**
   - Current: Separate files for x86 and ARM
   - Alternative: Single template with architecture detection
   - **Recommendation**: Keep separate (clearer, easier to test)

3. **Deprecation timeline for old templates?**
   - Option A: 2 months (faster cleanup)
   - Option B: 6 months (safer migration period)
   - **Recommendation**: 3 months (balance stability and progress)

4. **Should deprecated templates be deleted or archived?**
   - Option A: Move to `templates/archived/` (keep for reference)
   - Option B: Delete (cleaner codebase)
   - **Recommendation**: Archive (useful for troubleshooting old instances)

## Appendix: Example Template Hierarchy

### R Research Family (Current - Already Good ✅)

```
Ubuntu 24.04 LTS (x86_64)                [Tier 1: Base OS]
    └─> R Base (Ubuntu 24.04)            [Tier 3: Language]
            ├─> R + RStudio Server       [Tier 3: + IDE]
            │      └─> R Publishing Stack [Tier 4: Domain]
            └─> R Research Full Stack    [Tier 4: Monolithic]
```

### Python ML Family (Proposed - To Create 🆕)

```
Ubuntu 24.04 LTS (x86_64)                [Tier 1: Base OS]
    └─> Ubuntu 24.04 + Conda (x86_64)    [Tier 2: Package Manager]
            └─> Python Base (Conda)      [Tier 3: Language]
                    ├─> Python + Jupyter [Tier 3: + Notebooks]
                    │      └─> Python ML Stack [Tier 4: Domain]
                    └─> Python ML Configurable [Tier 4: Monolithic]
```

### Multi-Language Family (Proposed - To Create 🆕)

```
Ubuntu 24.04 + Conda (x86_64)            [Tier 2: Package Manager]
    ├─> Python Base (Conda)              [Tier 3: Python]
    │      └─> Data Science Minimal     [Tier 4: Python + R basics]
    └─> R Base (Conda)                   [Tier 3: R]
           └─> Data Science Full        [Tier 4: Python + R + Julia]
```

---

**Status**: Proposal - Ready for Review
**Next Steps**:
1. Get feedback on proposed organization
2. Prioritize which templates to refactor first
3. Begin implementation (Week 1: Foundation templates)
