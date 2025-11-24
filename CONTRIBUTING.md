# Contributing to Prism

Thank you for your interest in contributing to Prism! We welcome contributions from the community and are grateful for any help you can provide.

## Table of Contents

- [Before You Start](#before-you-start)
- [Ways to Contribute](#ways-to-contribute)
- [What Contributors CAN Do](#what-contributors-can-do)
- [What Contributors CANNOT Do](#what-contributors-cannot-do)
- [Development Setup](#development-setup)
- [Pull Request Requirements](#pull-request-requirements)
- [Testing Requirements](#testing-requirements)
- [Code Quality Standards](#code-quality-standards)
- [Security Considerations](#security-considerations)
- [Documentation Requirements](#documentation-requirements)
- [Plugin/Extension Development](#pluginextension-development)
- [Issue Labels](#issue-labels)
- [Code of Conduct](#code-of-conduct)
- [License](#license)

## Before You Start

1. **Search existing issues** before opening a new one to avoid duplicates
2. **Open an issue first** for bugs, feature requests, or significant changes
3. **Check persona alignment** - Does your contribution improve one of the [5 persona workflows](docs/USER_SCENARIOS/)?
4. **Review design principles** - Ensure changes follow [DESIGN_PRINCIPLES.md](docs/DESIGN_PRINCIPLES.md)
5. **Read the UX evaluation** - Check if your contribution addresses known [UX issues](docs/architecture/UX_EVALUATION_AND_RECOMMENDATIONS.md)

## Ways to Contribute

- 🐛 **Bug Reports**: Submit detailed bug reports with reproduction steps
- ✨ **Feature Requests**: Suggest new features aligned with user scenarios
- 📝 **Documentation**: Improve guides, fix typos, add examples
- 🧪 **Testing**: Add test coverage, report test failures
- 🎨 **Templates**: Propose new research environment templates
- 🔌 **Plugins**: Develop plugins/extensions for daemon functionality (see [Plugin Development](#pluginextension-development))
- 💻 **Code**: Submit PRs for issues labeled `help wanted` or `good first issue`

## What Contributors CAN Do

- ✅ Submit PRs for issues with `help wanted` or `good first issue` labels
- ✅ Propose new research environment templates
- ✅ Improve documentation and user guides
- ✅ Report bugs with detailed reproduction steps
- ✅ Suggest features aligned with the 5 persona scenarios
- ✅ Develop plugins/extensions (Phase 6 - see [Plugin Development](#pluginextension-development))
- ✅ Ask questions and request clarification on issues
- ✅ Review other contributors' PRs (helpful but not required)

## What Contributors CANNOT Do

- ❌ Open PRs without the `help wanted` label or explicit acceptance criteria
- ❌ Submit PRs for issues marked `core` (reserved for maintainers)
- ❌ Expand PR scope beyond what the issue describes
- ❌ Break multi-modal parity (changes must work across CLI/TUI/GUI)
- ❌ Merge without tests passing and maintainer approval
- ❌ Commit AWS credentials, secrets, or sensitive data

## Development Setup

### Prerequisites

- Go 1.24.4 or later
- Node.js 18+ (for GUI development)
- AWS CLI configured with credentials
- Git
- Make

### Getting Started

```bash
# Clone the repository
git clone https://github.com/scttfrdmn/prism.git
cd prism

# Build all components
make build

# Run tests
make test

# Build specific components
go build -o bin/prism ./cmd/prism/       # CLI
go build -o bin/prismd ./cmd/prismd/     # Daemon
go build -o bin/prism-gui ./cmd/prism-gui/  # GUI
```

See [Development Setup Guide](docs/development/DEVELOPMENT_SETUP.md) for detailed instructions.

## Pull Request Requirements

### Before Submitting

1. **Create a feature branch**: `git checkout -b my-feature-name`
2. **Make your changes**: Follow code quality standards below
3. **Add tests**: All new functionality requires tests
4. **Run tests**: Ensure all tests pass (`make test`)
5. **Update documentation**: Update relevant guides if user-facing
6. **Check multi-modal parity**: Ensure CLI/TUI/GUI consistency
7. **Commit your changes**: Use clear, descriptive commit messages

### Submitting

```bash
# Push your branch
git push origin my-feature-name

# Create PR using GitHub CLI
gh pr create --web

# Or create PR manually on GitHub
```

### PR Template

Your PR description should include:

- **Summary**: Brief description of changes (1-2 sentences)
- **Issue**: Link to related issue (`Fixes #123`)
- **Persona Impact**: Which user scenarios does this improve?
- **Testing**: How was this tested?
- **Multi-Modal**: CLI/TUI/GUI parity maintained?
- **Documentation**: Which docs were updated?
- **Screenshots**: For UI changes, include before/after

### PR Review Process

- Maintainers will review PRs within 1-2 weeks
- Address review feedback promptly
- Once approved, maintainers will merge
- Manual pages and release notes auto-generate at release time

## Testing Requirements

### Test Coverage

- **Unit tests**: Required for all new functions/methods
- **Integration tests**: Required for AWS operations
- **Manual testing**: Required for persona workflows when applicable
- **Test must pass**: All tests must pass before merge

### Running Tests

```bash
# Run all tests
make test

# Run specific package tests
go test ./pkg/aws/...

# Run with coverage
go test -cover ./...

# Run with race detection
go test -race ./...
```

### AWS Test Guidelines

- Use AWS mocking for unit tests (no real AWS calls)
- Integration tests clearly marked and documented
- Never commit AWS credentials in tests
- Use environment variable overrides for testing

## Code Quality Standards

### Go Code Style

- Follow standard Go formatting (`go fmt`, `gofmt`)
- Run linters (`golangci-lint run`)
- Keep cyclomatic complexity reasonable (`gocyclo`)
- Write clear, self-documenting code
- Add comments for complex logic

### Code Organization

- Follow existing package structure
- Separate concerns (daemon, API, CLI, TUI, GUI)
- Use interfaces for testability
- Keep files focused and reasonably sized

### Multi-Modal Parity

**CRITICAL**: Features must work across all interfaces:

- **CLI**: Command-line interface (`internal/cli/`)
- **TUI**: Terminal UI (`internal/tui/`)
- **GUI**: Graphical interface (`cmd/prism-gui/`)

Changes to one interface often require updates to others. Test all three!

## Security Considerations

### Security Requirements

- ✅ **No credentials in code**: Never commit AWS credentials, API keys, or secrets
- ✅ **Input validation**: Validate all user input
- ✅ **Error handling**: Handle errors securely (don't expose sensitive info)
- ✅ **AWS best practices**: Follow AWS security guidelines
- ✅ **OWASP awareness**: Avoid top 10 vulnerabilities

### Reporting Security Issues

**DO NOT** open public issues for security vulnerabilities. Instead:

1. Email security concerns to the maintainer privately
2. Include detailed description and reproduction steps
3. Allow reasonable time for fix before public disclosure

## Documentation Requirements

### When to Update Docs

- **User-facing features**: Update appropriate [user guides](docs/user-guides/)
- **Admin features**: Update [admin guides](docs/admin-guides/)
- **Architecture changes**: Update [architecture docs](docs/architecture/)
- **Breaking changes**: Update ROADMAP.md and CHANGELOG
- **Template changes**: Update template documentation

### Documentation Standards

- Clear, concise language
- Include code examples where applicable
- Update table of contents if adding sections
- Test all command examples before submitting
- Follow existing documentation style

## Plugin/Extension Development

### Overview

**Status**: Planned for **Phase 6 (v0.7.0 - Q4 2026)**

Prism will support daemon plugins/extensions via a **gRPC-based architecture** allowing custom functionality without forking the codebase.

### Plugin Use Cases

- **Authentication Providers**: LDAP, Shibboleth, CAS, OAuth2, SAML
- **HPC Integration**: SLURM, PBS, LSF job submission
- **Cost Integration**: Custom billing/chargeback systems
- **Audit Logging**: Compliance requirements (HIPAA, NIST 800-171)
- **Analytics**: Usage tracking, institutional reporting
- **Third-Party Services**: GitLab, Jira, ServiceNow integration

### Plugin Architecture

Plugins run as **separate processes** communicating via gRPC:

```
┌─────────────┐         gRPC        ┌─────────────┐
│   Prismd    │◄───────────────────►│   Plugin    │
│   Daemon    │                     │   Process   │
└─────────────┘                     └─────────────┘
```

**Benefits**:
- ✅ Security isolation (plugin crash doesn't kill daemon)
- ✅ Credential sandboxing (controlled AWS access)
- ✅ Version independence (plugin and daemon update separately)
- ✅ Language agnostic (plugins in any language)

### Contributing Plugins

**Path to Contributing Plugins**:

1. **Propose Plugin** (Phase 6): Open GitHub issue describing:
   - Plugin purpose and use case
   - Required permissions (AWS, filesystem, network)
   - Hook points needed (pre-launch, post-launch, auth, etc.)
   - Target user persona

2. **Review & Design**: Maintainers review and provide feedback on:
   - Security implications
   - Permission model appropriateness
   - Hook point availability
   - API stability commitments

3. **Development**: Once approved:
   - Use Go or Python SDK (provided in Phase 6)
   - Follow plugin development guide
   - Include tests and documentation
   - Submit as community plugin

4. **Plugin Registry**: Community plugins hosted in:
   - Personal GitHub repositories (linked in registry)
   - Institutional repositories (for internal plugins)
   - Optional official plugin repository (curated collection)

**Note**: Plugin system is not yet implemented. Early plugin proposals are welcome to help shape the design! Open an issue tagged `plugin-proposal`.

### Why Plugins vs PRs?

Plugins are appropriate when:
- ✅ Functionality is institution-specific (e.g., custom LDAP)
- ✅ Integration with proprietary systems (e.g., internal billing)
- ✅ Experimental features needing rapid iteration
- ✅ Compliance requirements specific to your organization
- ✅ Third-party service integrations

Standard PRs are appropriate when:
- ✅ Core functionality benefiting all users
- ✅ Bug fixes and performance improvements
- ✅ General research templates
- ✅ Improvements to personas workflows
- ✅ Documentation and examples

## Issue Labels

Understanding issue labels helps you find contribution opportunities:

- **`help wanted`**: Open for community contributions (you can work on these!)
- **`good first issue`**: Suitable for new contributors (great starting point)
- **`core`**: Reserved for maintainers only (do not submit PRs)
- **`bug`**: Something isn't working correctly
- **`enhancement`**: New feature or improvement
- **`documentation`**: Documentation improvements
- **`template-request`**: Request for new research template
- **`plugin-proposal`**: Proposal for future plugin/extension
- **`needs-design`**: Requires design proposal before implementation
- **`v0.5.x`**: Targeted for specific version

## Code of Conduct

All contributors must follow the [Code of Conduct](CODE_OF_CONDUCT.md). We are committed to providing a welcoming and inclusive environment for all.

**Quick Summary**:
- Be respectful and inclusive
- Welcome diverse perspectives
- Accept constructive criticism gracefully
- Focus on what's best for the community
- Show empathy towards other contributors

## License

By contributing to Prism, you agree that your contributions will be licensed under the **Apache License, Version 2.0**.

Contributions are submitted under the same terms and conditions as the project license. See [LICENSE](LICENSE) for details.

---

## Questions?

- **General Questions**: Open a [GitHub Discussion](https://github.com/scttfrdmn/prism/discussions)
- **Bug Reports**: Open a [GitHub Issue](https://github.com/scttfrdmn/prism/issues)
- **Feature Ideas**: Open an issue with `enhancement` label
- **Plugin Proposals**: Open an issue with `plugin-proposal` label
- **Security Issues**: Email maintainer privately (do not open public issue)

Thank you for contributing to Prism! 🎉
