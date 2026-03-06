$ErrorActionPreference = 'Stop'

$packageName = 'prism'
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"

# Test channel configuration - uncomment for test releases
$preRelease = $env:CHOCOLATEY_PRERELEASE -eq 'true'
$version = if ($preRelease) { '0.5.1-beta' } else { '0.5.1' }
$repoPath = if ($preRelease) { 'releases-dev' } else { 'releases' }

$url = "https://github.com/scttfrdmn/prism/$repoPath/download/v$version/cws-windows-amd64.zip"
$checksum = 'REPLACE_WITH_ACTUAL_CHECKSUM_AFTER_BUILDING'
$checksumType = 'sha256'

$packageArgs = @{
  packageName   = $packageName
  unzipLocation = $toolsDir
  url           = $url
  checksum      = $checksum
  checksumType  = $checksumType
}

Install-ChocolateyZipPackage @packageArgs

# Create shortcut
$startMenuPath = [Environment]::GetFolderPath('CommonStartMenu')
$shortcutPath = Join-Path $startMenuPath 'Programs\Prism\Prism.lnk'
$targetPath = Join-Path $toolsDir 'prism-gui.exe'

# Create directory if it doesn't exist
if (!(Test-Path (Split-Path $shortcutPath))) {
  New-Item -ItemType Directory -Path (Split-Path $shortcutPath) | Out-Null
}

# Create shortcut if GUI executable exists
if (Test-Path $targetPath) {
  Install-ChocolateyShortcut -ShortcutFilePath $shortcutPath -TargetPath $targetPath -Description 'Prism - Research environments in the cloud'
}

# Add to PATH
$binPath = Join-Path $toolsDir 'cws.exe'
$daemonPath = Join-Path $toolsDir 'prismd.exe'
Install-BinFile -Name 'cws' -Path $binPath
Install-BinFile -Name 'prismd' -Path $daemonPath

# Install Windows service for auto-startup
$serviceWrapperPath = Join-Path $toolsDir 'prism-service.exe'
if (Test-Path $serviceWrapperPath) {
    Write-Host "Installing Prism Windows service..."
    try {
        Start-Process -FilePath $serviceWrapperPath -ArgumentList 'install' -Wait -Verb RunAs
        Write-Host "✅ Prism service installed successfully"
        Write-Host "   Service will start automatically on system boot"
    }
    catch {
        Write-Warning "⚠️  Failed to install Windows service: $_"
        Write-Host "   You can manually install the service later with:"
        Write-Host "   $serviceWrapperPath install"
    }
} else {
    Write-Warning "⚠️  Windows service wrapper not found. Service auto-startup not configured."
}

Write-Host ""
Write-Host "🎉 Prism v$version has been installed!"
Write-Host ""
Write-Host "📦 Installed Components:"
Write-Host "  • CLI (cws) - Available in PATH"
Write-Host "  • Daemon (prismd) - Available in PATH"
if (Test-Path $targetPath) {
    Write-Host "  • GUI - Available in Start Menu"
}
if (Test-Path $serviceWrapperPath) {
    Write-Host "  • Windows Service - Auto-starts on boot"
}
Write-Host ""
Write-Host "🚀 Quick Start:"
Write-Host "  cws --help                    # Show CLI help"
Write-Host "  cws daemon status             # Check daemon status"
Write-Host ""
Write-Host "🔧 Service Management:"
Write-Host "  sc query PrismDaemon        # Check service status"
Write-Host "  sc start PrismDaemon        # Start service manually"
Write-Host "  sc stop PrismDaemon         # Stop service"