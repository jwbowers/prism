# Template Review Checklist

## Purpose

This checklist ensures consistent, high-quality community template submissions. Use this when reviewing template PRs to verify all requirements are met.

---

## Quick Review (5 Minutes)

Fast checks before deep review:

- [ ] **PR uses template submission form** (not generic PR template)
- [ ] **Template file in `templates/community/`**
- [ ] **Documentation file in `docs/user-guides/`**
- [ ] **CI passed** (green checkmark on PR)
- [ ] **No merge conflicts**

❌ If any quick checks fail, request changes immediately.

---

## File Structure (2 Minutes)

### Template File

- [ ] **Location**: `templates/community/[name].yml`
- [ ] **Naming**: lowercase-with-hyphens
- [ ] **YAML valid**: No syntax errors
- [ ] **Character encoding**: UTF-8

### Documentation File

- [ ] **Location**: `docs/user-guides/[NAME].md`
- [ ] **Naming**: UPPERCASE_WITH_UNDERSCORES.md
- [ ] **Markdown valid**: No broken links
- [ ] **Headers present**: Overview, Quick Start, Troubleshooting

---

## Required Fields (3 Minutes)

Verify these fields exist and are valid:

```yaml
# Core fields
- [ ] name: "Clear, descriptive name"
- [ ] description: "What this template provides"
- [ ] base_os: ubuntu-24.04 | ubuntu-22.04 | amazonlinux-2023
- [ ] architecture: x86_64 | arm64 | both
- [ ] package_manager: apt | yum | conda | spack

# Configuration
- [ ] packages: (at least one package)
- [ ] users: (at least one user)
- [ ] ports: [22, ...] (SSH + any web services)

# Metadata
- [ ] version: "X.Y.Z" (semantic versioning)
- [ ] tags.category: (from approved list)
```

### Optional But Recommended

- [ ] author: "Name <email>"
- [ ] services: (if web interface)
- [ ] idle_detection: (for cost optimization)
- [ ] post_install: (if custom setup needed)

---

## Content Review (10 Minutes)

### Template Quality

**Packages**:
- [ ] **No duplicates** across package lists
- [ ] **Latest stable versions** (or pinned with reason)
- [ ] **No proprietary software** (unless documented how to obtain)
- [ ] **Organized logically** (system, language-specific, tools)

**Users**:
- [ ] **Default user present** (researcher, ubuntu, etc.)
- [ ] **Sudo access appropriate** (not all users need sudo)
- [ ] **No hardcoded passwords**

**Services**:
- [ ] **Ports match services** (Jupyter=8888, RStudio=8787, etc.)
- [ ] **Start commands valid** (tested)
- [ ] **Check commands present** (for health monitoring)

**Post-Install Script**:
- [ ] **Uses `set -e`** (fails on error)
- [ ] **Idempotent** (can run multiple times safely)
- [ ] **No hardcoded paths** to user home directories
- [ ] **Progress messages** (echo statements)
- [ ] **Cleans up temp files**

### Documentation Quality

**Overview**:
- [ ] **Clear description** (what, why, who for)
- [ ] **Target audience** identified
- [ ] **Prerequisites listed** (if any)

**Quick Start**:
- [ ] **Launch command** (`prism launch ...`)
- [ ] **Connect command** (`prism connect ...`)
- [ ] **First steps** (what to do after connecting)
- [ ] **Estimated time** (5 minutes? 30 minutes?)

**What's Included**:
- [ ] **Software versions** listed
- [ ] **Key packages** highlighted
- [ ] **Services and ports** documented

**Troubleshooting**:
- [ ] **Common issues** listed (at least 2-3)
- [ ] **Solutions provided** (not just problems)
- [ ] **Where to get help** (GitHub, docs, community)

---

## Security Review (5 Minutes)

### Immediate Rejection Criteria

❌ **Reject immediately if found**:
- [ ] Hardcoded API keys or secrets
- [ ] Passwords in plaintext
- [ ] Private SSH keys
- [ ] AWS credentials
- [ ] Database connection strings with passwords
- [ ] Proprietary software binaries (without license)

### Security Best Practices

✅ **Should be present**:
- [ ] **Principle of least privilege** (minimal sudo usage)
- [ ] **Latest security patches** (system updates in post_install)
- [ ] **No disabled firewalls** or security features
- [ ] **SSL/TLS** for web services (at least documented)

### Security Warnings

⚠️ **Flag for discussion**:
- [ ] Installs from non-official sources (PPAs, custom repos)
- [ ] Downloads binaries without checksum verification
- [ ] Disables SELinux or AppArmor
- [ ] Opens unusual ports
- [ ] Uses latest/edge package versions (may be unstable)

---

## Testing (15-30 Minutes)

### Automated Tests

- [ ] **`prism templates validate`** passes
- [ ] **CI build** succeeds
- [ ] **No linting errors**

### Manual Testing

**Test 1: x86_64 Launch** (Required)
```bash
# Launch in test region
AWS_REGION=us-east-1 prism launch [template-name] test-x86

# Verify:
- [ ] Instance launches successfully
- [ ] All packages installed
- [ ] Services running (check ports)
- [ ] Post-install completed without errors
- [ ] Can connect via SSH

# Clean up
prism delete test-x86
```

**Test 2: ARM64 Launch** (If template claims support)
```bash
AWS_REGION=us-east-1 prism launch --architecture arm64 [template-name] test-arm

# Verify same as x86_64

prism delete test-arm
```

**Test 3: Multi-Region** (Recommended)
```bash
# Test in second region
AWS_REGION=us-west-2 prism launch [template-name] test-west

# Verify same as first region

prism delete test-west
```

### Service Testing

For templates with web interfaces:

- [ ] **Service starts** automatically
- [ ] **Port accessible** (curl or browser check)
- [ ] **Authentication works** (if configured)
- [ ] **Basic functionality** (can create notebook, run code, etc.)

### Performance Check

- [ ] **Launch time reasonable** (< 10 minutes)
- [ ] **No excessive downloads** (> 10GB)
- [ ] **Post-install < 30 minutes**

---

## Categorization (2 Minutes)

### Category Validation

- [ ] **Category matches content**:
  - Machine Learning → TensorFlow, PyTorch, scikit-learn
  - Bioinformatics → BLAST, bowtie2, bedtools
  - Data Science → pandas, R, tidyverse
  - Geospatial → QGIS, GDAL, PostGIS
  - (etc.)

- [ ] **Level appropriate**:
  - Beginner: < 10 packages, simple setup
  - Intermediate: 10-30 packages, some config
  - Advanced: 30+ packages or complex setup

### Tags

- [ ] **Type tag** present: research | teaching | development
- [ ] **Additional tags** helpful: gpu, gui, jupyter, rstudio, etc.

---

## Multi-Region Compatibility (3 Minutes)

### Regional Considerations

- [ ] **No region-specific AMIs** (uses base_os field correctly)
- [ ] **No hardcoded endpoints** (use AWS service discovery)
- [ ] **Documentation mentions** regional limitations (if any)

### Architecture Support

If template claims `architecture: both`:
- [ ] **Tested on x86_64** ✅
- [ ] **Tested on ARM64** ✅
- [ ] **No arch-specific packages** without conditionals

---

## Versioning & Updates (2 Minutes)

### Version Number

- [ ] **Semantic versioning**: X.Y.Z
- [ ] **Version 1.0.0** for new templates
- [ ] **Version increment** for updates (1.0.0 → 1.0.1 or 1.1.0)

### Changelog

For template updates:
- [ ] **PR description** explains changes
- [ ] **Breaking changes** flagged clearly
- [ ] **Migration guide** (if breaking changes)

---

## Community Standards (2 Minutes)

### Code of Conduct

- [ ] **Respectful language** in all documentation
- [ ] **Inclusive examples** (not assuming gender, location, expertise)
- [ ] **No offensive content**

### Attribution

- [ ] **Author credited** properly
- [ ] **Dependencies acknowledged** (if using others' work)
- [ ] **License compatible** with project (Apache 2.0)

---

## Final Approval Checklist

Before merging:

- [ ] ✅ **All required fields** present
- [ ] ✅ **Security review** passed
- [ ] ✅ **Manual testing** completed (x86_64 minimum)
- [ ] ✅ **Documentation** complete and clear
- [ ] ✅ **CI** green
- [ ] ✅ **No merge conflicts**
- [ ] ✅ **Author responded** to all feedback
- [ ] ✅ **Squash commits** (if multiple commits)

### Merge Message

Use this format:

```
feat(templates): Add [Template Name] (#PR_NUMBER)

[Template Name] provides [brief description].

Features:
- [Key feature 1]
- [Key feature 2]
- [Key feature 3]

Tested on: x86_64 in us-east-1, us-west-2
Category: [Category Name]
Author: @github_username

Closes #PR_NUMBER
```

---

## Rejection Reasons

Common reasons to request changes:

1. **Security issues** (hardcoded secrets)
2. **Missing documentation** (no quick start)
3. **Testing incomplete** (no manual test)
4. **Required fields missing** (no version, category)
5. **Template doesn't validate** (CI failures)
6. **Too broad scope** (mixing unrelated tools)
7. **Duplicate** of existing template
8. **No clear use case** (what problem does it solve?)

---

## Review Time Estimates

- **Simple template** (10-20 packages): 20-30 minutes
- **Medium template** (20-50 packages): 30-45 minutes
- **Complex template** (50+ packages, inheritance): 45-60 minutes

---

## Tips for Efficient Reviews

1. **Use PR template** to ensure all info provided
2. **Run CI first** before manual testing
3. **Test in parallel** (x86 + ARM at same time)
4. **Check existing templates** for similar functionality
5. **Provide clear feedback** with examples
6. **Be encouraging** - community contributors are volunteers!

---

## Getting Help

**Not sure about something?**
- Tag another maintainer
- Ask in maintainer channel
- Consult security team for security concerns
- Check existing template reviews for precedent

**Remember**: We're here to help contributors succeed! 🎉

---

## Related Documentation

- [Community Template Guide](COMMUNITY_TEMPLATE_GUIDE.md) - Contributor guide
- [Template README](../../templates/README.md) - Template structure
- [Security Guidelines](../admin-guides/SECURITY_COMPLIANCE_ROADMAP.md) - Security standards

---

**Last Updated**: January 2026 | **Version**: 1.0
