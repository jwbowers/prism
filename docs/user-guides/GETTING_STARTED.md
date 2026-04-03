# Installation & Setup

## Install Prism

=== "macOS (Homebrew)"
    ```bash
    brew install scttfrdmn/tap/prism
    ```

=== "Windows (Scoop)"
    ```powershell
    scoop bucket add scttfrdmn https://github.com/scttfrdmn/scoop-bucket
    scoop install prism
    ```

=== "Linux"
    Download the latest tarball from the [releases page](https://github.com/scttfrdmn/prism/releases/latest), extract, and add to your PATH.

Verify:
```bash
prism version
```

---

## Connect AWS credentials

Prism uses your AWS credentials to launch instances in your account.

**Step 1 — Configure the AWS CLI** (skip if already done):
```bash
aws configure
# Prompts for: AWS Access Key ID, Secret Access Key, region, output format
```

Verify it works:
```bash
aws sts get-caller-identity
```

**Step 2 — Add a Prism profile**:
```bash
prism profile add
```

This interactive wizard links a Prism profile name to your AWS credentials and default region. You only need to do this once.

For detailed IAM permission requirements, see the [Administrator Guide](../admin-guides/ADMINISTRATOR_GUIDE.md).

---

## First-time setup wizard

If this is your first time running Prism, the init wizard covers AWS setup, profile creation, and a test launch:

```bash
prism init
```

---

## Launch your first workspace

```bash
# See what's available
prism templates

# Launch a Python + Jupyter environment
prism workspace launch python-ml my-first-project

# Check it's ready (takes ~2 minutes)
prism workspace list

# Get connection info
prism workspace connect my-first-project
```

`prism workspace connect` prints the SSH command and any web service URLs (Jupyter, RStudio, etc.).

---

## Two interfaces

### CLI
The `prism` command is the primary interface. It's scriptable and works in automated pipelines.

```bash
prism workspace launch python-ml my-project --size L --spot
```

### GUI (desktop app)
A visual desktop app for managing workspaces, storage, and settings.

```bash
prism gui
```

Or launch Prism from your Applications folder (macOS) / Start menu (Windows).

---

## Common workflows

### Data science
```bash
prism workspace launch python-ml data-analysis --size L
prism workspace connect data-analysis
# Jupyter is available at http://localhost:8888 via SSH tunnel
```

### R / statistics
```bash
prism workspace launch r-research stats-project
prism workspace connect stats-project
# RStudio Server available at http://localhost:8787 via SSH tunnel
```

### Shared storage
```bash
prism volume create shared-datasets
prism volume attach shared-datasets my-project
```

---

## Cost management

Stop workspaces when not in use — stopped instances have no compute cost:

```bash
prism workspace stop my-project
prism workspace start my-project   # Resume later
```

Hibernation preserves RAM state and reduces cost further:

```bash
prism workspace hibernate my-project
prism workspace resume my-project
```

---

## Troubleshooting

**"Daemon not running"**
```bash
prism admin daemon status
prism admin daemon stop     # Then run any prism command to auto-restart
```

**"AWS credentials not found"**
```bash
aws sts get-caller-identity    # Verify credentials work
aws configure                  # Reconfigure if needed
```

**Instance launch fails**
```bash
prism workspace launch python-ml my-project --region us-east-1
```

For more, see the [Troubleshooting Guide](TROUBLESHOOTING.md).

---

## Next steps

- **[Quick Start](QUICK_START.md)** — 5-minute guide
- **[CLI Reference](CLI_REFERENCE.md)** — all commands
- **[GUI Guide](GUI_USER_GUIDE.md)** — desktop app
- **[Templates](TEMPLATE_FORMAT.md)** — template format and customization
