# macOS Installation Guide

Prism can be installed on macOS via Homebrew for a streamlined command-line and GUI experience.

## Quick Start

```bash
# Install via Homebrew
brew tap scttfrdmn/prism
brew install prism

# Verify installation
prism version
```

## Installation Methods

### Method 1: Homebrew (Recommended)

**Best for:** Most macOS users, provides automatic updates and easy management.

```bash
# Add Prism tap
brew tap scttfrdmn/prism

# Install Prism
brew install prism

# Verify installation
prism version
prismd version
```

**Includes:**
- `prism` CLI tool
- `prismd` daemon
- `prism-gui` desktop application (if GUI support is available)
- Automatic PATH configuration
- Easy updates via `brew upgrade`

### Method 2: DMG Installer

**Best for:** Users who prefer a graphical installer or want the Prism desktop GUI.

1. Go to the [Prism Releases page](https://github.com/scttfrdmn/prism/releases/latest)
2. Download `Prism-v*-macOS.dmg` (universal binary, works on Intel and Apple Silicon)
3. Open the `.dmg` file
4. Drag **Prism.app** to your **Applications** folder
5. Eject the disk image

**First launch (Gatekeeper):**

macOS may block the app since it is not notarized yet. To open it:

1. In Finder, navigate to Applications
2. Right-click **Prism.app** → **Open**
3. Click **Open** in the security dialog (only required once)

Or via Terminal:

```bash
xattr -d com.apple.quarantine /Applications/Prism.app
```

**Install CLI tools from the app bundle (optional):**

```bash
# Install prism and prismd to /usr/local/bin
sudo cp /Applications/Prism.app/Contents/MacOS/prism /usr/local/bin/prism
sudo cp /Applications/Prism.app/Contents/MacOS/prismd /usr/local/bin/prismd
sudo chmod +x /usr/local/bin/prism /usr/local/bin/prismd

# Verify
prism --help
```

### Method 3: Direct Binary Download

**Best for:** Users who prefer manual installation or don't use Homebrew.

```bash
# Intel Macs
curl -L https://github.com/scttfrdmn/prism/releases/latest/download/prism-darwin-amd64.tar.gz | tar xz

# Apple Silicon Macs
curl -L https://github.com/scttfrdmn/prism/releases/latest/download/prism-darwin-arm64.tar.gz | tar xz

# Move binaries to PATH
sudo mv prism prismd /usr/local/bin/
```

## Post-Installation Setup

### 1. Configure AWS Credentials

Prism requires AWS credentials to manage cloud resources:

```bash
# Install AWS CLI if needed
brew install awscli

# Configure your AWS credentials
aws configure
# Enter your AWS Access Key ID
# Enter your AWS Secret Access Key
# Enter your default region (e.g., us-west-2)
# Enter output format (json)
```

### 2. Verify Installation

```bash
# Check Prism version
prism version

# List available templates
prism templates

# Check daemon status (auto-starts as needed)
prism admin daemon status
```

## Using Prism on macOS

### Launch GUI Application

```bash
# Start the desktop application
prism gui
```

### Use Command Line

```bash
# See all available commands
prism --help

# Launch your first environment
prism workspace launch python-ml my-first-project
```

## Updating Prism

```bash
# Update via Homebrew
brew update
brew upgrade prism

# Verify new version
prism version
```

## Troubleshooting

### "Command not found: prism"

```bash
# Check if Prism is installed
brew list prism

# Reinstall if needed
brew reinstall prism

# Verify PATH includes Homebrew bin
echo $PATH | grep homebrew
```

### Permission Issues

```bash
# Ensure proper permissions on binaries
ls -la $(which prism)
ls -la $(which prismd)

# Fix permissions if needed
brew reinstall prism
```

### Daemon Connection Issues

```bash
# Check daemon status (daemon auto-starts as needed)
prism admin daemon status

# Stop daemon and let it auto-restart on next command
prism admin daemon stop
prism templates
```

## Uninstalling

```bash
# Remove Prism via Homebrew
brew uninstall prism
brew untap scttfrdmn/prism

# Remove configuration (optional)
rm -rf ~/.prism
```

## Next Steps

- See [Getting Started Guide](GETTING_STARTED.md) for first-time usage
- Explore [Template Format](TEMPLATE_FORMAT.md) to create custom environments
- Check [Troubleshooting](TROUBLESHOOTING.md) for common issues
