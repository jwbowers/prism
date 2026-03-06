$ErrorActionPreference = 'Stop'

$packageName = 'prism'
$toolsDir = "$(Split-Path -parent $MyInvocation.MyCommand.Definition)"

# Uninstall Windows service first
$serviceWrapperPath = Join-Path $toolsDir 'prism-service.exe'
if (Test-Path $serviceWrapperPath) {
    Write-Host "Uninstalling Prism Windows service..."
    try {
        Start-Process -FilePath $serviceWrapperPath -ArgumentList 'remove' -Wait -Verb RunAs
        Write-Host "✅ Prism service uninstalled successfully"
    }
    catch {
        Write-Warning "⚠️  Failed to uninstall Windows service: $_"
        Write-Host "   You may need to manually remove the service with:"
        Write-Host "   sc delete PrismDaemon"
    }
}

# Remove shortcut
$startMenuPath = [Environment]::GetFolderPath('CommonStartMenu')
$shortcutPath = Join-Path $startMenuPath 'Programs\Prism\Prism.lnk'

if (Test-Path $shortcutPath) {
  Remove-Item $shortcutPath -Force
}

# Try to remove the shortcut directory if empty
$shortcutDir = Split-Path $shortcutPath
if (Test-Path $shortcutDir) {
  if ((Get-ChildItem $shortcutDir | Measure-Object).Count -eq 0) {
    Remove-Item $shortcutDir -Force
  }
}

# Remove from PATH
Uninstall-BinFile -Name 'cws'
Uninstall-BinFile -Name 'prismd'

Write-Host ""
Write-Host "✅ Prism has been uninstalled."
Write-Host ""
Write-Host "📋 What was removed:"
Write-Host "  • CLI and daemon binaries"
Write-Host "  • Start Menu shortcuts"
Write-Host "  • Windows Service (auto-startup disabled)"
Write-Host ""
Write-Host "📁 Configuration and data preserved in:"
Write-Host "  • %USERPROFILE%\.prism\"
Write-Host "  • %PROGRAMDATA%\Prism\"