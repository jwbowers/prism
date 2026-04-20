# Windows Installation Guide

Prism can be installed on Windows via MSI installer or direct binary download.

## Installation Methods

### Method 1: MSI Installer (Recommended)

**Best for:** Most Windows users — installs all components, configures PATH, and sets up the Prism background service.

1. Go to the [Prism Releases page](https://github.com/scttfrdmn/prism/releases/latest)
2. Download `Prism-v*-x64.msi`
3. Double-click the `.msi` file to start the installer
4. Follow the setup wizard (accept defaults unless you need a custom install directory)
5. Click **Finish**

The installer will:
- Install `prism.exe`, `prismd.exe`, and `prism-gui.exe`
- Add the install directory to your `PATH`
- Register **PrismDaemon** as a Windows Service (auto-starts on boot)
- Create a Start Menu shortcut for the Prism GUI

**Verify installation:**

```powershell
prism --help
prism admin daemon status
```

#### Silent Install

For automated or enterprise deployments:

```powershell
# Silent install with default options
msiexec /i Prism-v0.35.3-x64.msi /quiet

# Silent install to a custom directory
msiexec /i Prism-v0.35.3-x64.msi /quiet INSTALLDIR="C:\Tools\Prism"

# Silent install with a log file
msiexec /i Prism-v0.35.3-x64.msi /quiet /log install.log
```

#### Uninstall

```powershell
# Via Add/Remove Programs (recommended)
# Settings → Apps → Search "Prism" → Uninstall

# Silent uninstall
msiexec /x Prism-v0.35.3-x64.msi /quiet
```

### Method 2: Direct Binary Download

**Best for:** Users without administrator rights or who prefer a portable installation.

```powershell
# Download and extract (PowerShell)
Invoke-WebRequest -Uri https://github.com/scttfrdmn/prism/releases/latest/download/prism-windows-amd64.zip -OutFile prism.zip
Expand-Archive prism.zip -DestinationPath $env:USERPROFILE\bin\prism

# Add to PATH (current session)
$env:PATH += ";$env:USERPROFILE\bin\prism"

# Add to PATH permanently
[Environment]::SetEnvironmentVariable("PATH", $env:PATH + ";$env:USERPROFILE\bin\prism", "User")
```

## Post-Installation Setup

### 1. Configure AWS Credentials

Prism requires AWS credentials to manage cloud resources:

```powershell
# Install AWS CLI
winget install Amazon.AWSCLI
# or download from https://aws.amazon.com/cli/

# Configure credentials
aws login
# Enter your AWS Access Key ID
# Enter your AWS Secret Access Key
# Enter your default region (e.g., us-west-2)
# Enter output format (json)
```

### 2. Verify Installation

```powershell
# Check Prism version
prism version

# List available templates
prism templates

# Check daemon status
prism admin daemon status
```

### 3. First Launch

```powershell
# Launch your first environment
prism workspace launch python-ml my-first-project

# Or open the desktop GUI
prism-gui
# (or use the Start Menu shortcut)
```

## Windows Service (PrismDaemon)

The MSI installer registers `prismd` as a Windows Service named **PrismDaemon** that starts automatically on boot.

```powershell
# Check service status
sc query PrismDaemon

# Start service manually (if stopped)
sc start PrismDaemon

# Stop service
sc stop PrismDaemon

# Disable auto-start (if you prefer manual control)
sc config PrismDaemon start= demand
```

When the service is not running, `prism` will auto-start the daemon as a background process on first use.

## Troubleshooting

### "prism is not recognized as a command"

The installer adds Prism to your PATH, but you may need to restart your terminal.

```powershell
# Open a new PowerShell or Command Prompt window, then:
prism --help

# If still not found, check PATH manually:
[Environment]::GetEnvironmentVariable("PATH", "Machine") | Select-String "Prism"

# Add manually if missing:
[Environment]::SetEnvironmentVariable("PATH", $env:PATH + ";C:\Program Files\Prism", "Machine")
```

### Windows Service Not Starting

```powershell
# Check Windows Event Log for errors
Get-EventLog -LogName Application -Source "PrismDaemon" -Newest 10

# Try starting manually
sc start PrismDaemon

# Or run prismd directly (bypasses service)
prismd
```

### Antivirus Blocking Prism

Some antivirus software may flag unsigned binaries. If Prism is quarantined:

1. Open your antivirus software
2. Add `C:\Program Files\Prism\` to the exclusions/allowlist
3. Restore quarantined files if needed

Code signing certificates are planned for a future release.

### GUI Doesn't Launch

```powershell
# Start GUI from command line to see error output
prism-gui

# Check if daemon is running (GUI requires it)
prism admin daemon status
prism admin daemon start
```

## Updating Prism

Download and run the new MSI installer — it will upgrade in place. The Windows Service will be restarted automatically.

For silent upgrades:

```powershell
msiexec /i Prism-v0.35.3-x64.msi /quiet
```

## Uninstalling

```powershell
# Via Settings (recommended)
# Settings → Apps → Apps & Features → Search "Prism" → Uninstall

# Silent uninstall
msiexec /x Prism-v0.35.3-x64.msi /quiet

# Remove configuration data (optional)
Remove-Item -Recurse -Force "$env:USERPROFILE\.prism"
Remove-Item -Recurse -Force "$env:PROGRAMDATA\Prism"
```

## Next Steps

- See [Getting Started Guide](QUICK_START.md) for first-time usage
- Read the [CLI User Guide](USER_GUIDE_RESEARCH_USERS.md) for command reference
- Check [Troubleshooting](TROUBLESHOOTING.md) for common issues
