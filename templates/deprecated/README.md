# Deprecated Templates

This directory contains templates that have been deprecated as part of the template library cleanup (Issue #429).

## Why These Templates Are Deprecated

As of January 2026, Prism's built-in template library has been simplified to contain only **base OS templates**. Application-specific templates should be created as:

1. **Community Templates** (`templates/community/`) - Contributed by the community
2. **Custom User Templates** (`~/.prism/templates/`) - Created by individual users
3. **Institutional Templates** - Hosted by organizations

## What's In This Directory

These templates represent the previous approach where Prism shipped with many application-specific templates (Python ML, R Research, QGIS, MATLAB, etc.). While functional, this approach had several issues:

- Hard to maintain (18+ templates to keep updated)
- Mixed concerns (OS configuration + application setup)
- Limited flexibility for users
- Difficult to test comprehensively

## Migration Path

### For Template Users

If you were using one of these templates, you have several options:

1. **Use the new base templates** + install applications manually:
   ```bash
   prism launch ubuntu-24-04-x86 my-instance
   prism connect my-instance
   # Install your applications using apt/conda/pip
   ```

2. **Create a custom template** based on the deprecated one:
   ```bash
   cp templates/deprecated/python-ml-workstation.yml ~/.prism/templates/my-python-ml.yml
   # Edit and customize as needed
   prism launch my-python-ml my-instance
   ```

3. **Use the R Research templates** (Issue #430):
   - New comprehensive R research templates are being developed
   - Will include RStudio Server, Quarto, LaTeX, and essential tools
   - Location: `templates/community/r-research-full-stack.yml`

### For Template Authors

If you want to convert a deprecated template to a community template:

1. Move it to `templates/community/`
2. Add proper metadata (author, version, etc.)
3. Make it inherit from a base template (recommended)
4. Test thoroughly
5. Submit a pull request

See `docs/development/COMMUNITY_TEMPLATE_GUIDE.md` for detailed instructions (Issue #434).

## Template Inheritance

The new recommended approach is to use **template inheritance**:

```yaml
name: "My Python ML Environment"
inherits: ["Ubuntu 24.04 LTS (x86_64)"]

packages:
  conda:
    - "python=3.12"
    - "numpy"
    - "pandas"
    # ...
```

This gives you:
- Cleaner templates (focus on your additions)
- Better maintainability (base OS updates propagate)
- Composition over duplication

## Removal Timeline

These deprecated templates will remain in this directory for **6 months** (until July 2026) to allow users to migrate. After that, they will be removed from the main repository.

They will remain available in git history and can be retrieved if needed.

## Questions?

- For questions about the template cleanup: See Issue #429
- For community template contributions: See Issue #434
- For R research templates: See Issues #430 and #431
- For general template questions: Check `docs/user-guides/TEMPLATE_FORMAT.md`
