# Prism Windows Service Manager
# PowerShell script for managing Prism daemon as a Windows service

param(
    [Parameter(Position = 0, Mandatory = $true)]
    [ValidateSet("install", "uninstall", "start", "stop", "restart", "status", "logs", "validate", "help")]
    [string]$Command,
    
    [switch]$Force,
    [switch]$Verbose,
    [switch]$WhatIf
)

# Service configuration
$ServiceName = "PrismDaemon"
$DisplayName = "Prism Daemon"
$Description = "Enterprise research management platform daemon for launching cloud research environments"
$ServiceExecutable = "prism-service.exe"

# Paths
$InstallPath = "${env:ProgramFiles}\Prism"
$ConfigPath = "${env:ProgramData}\Prism"
$LogPath = "${env:ProgramData}\Prism\Logs"
$ServicePath = Join-Path $InstallPath $ServiceExecutable

# Color output functions
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    
    $colors = @{
        "Red" = [ConsoleColor]::Red
        "Green" = [ConsoleColor]::Green  
        "Yellow" = [ConsoleColor]::Yellow
        "Blue" = [ConsoleColor]::Blue
        "Cyan" = [ConsoleColor]::Cyan
        "Magenta" = [ConsoleColor]::Magenta
        "White" = [ConsoleColor]::White
    }
    
    Write-Host $Message -ForegroundColor $colors[$Color]
}

function Write-Log {
    param([string]$Message)
    $timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
    Write-ColorOutput "[$timestamp] $Message" "White"
}

function Write-Success {
    param([string]$Message)
    Write-ColorOutput "✅ $Message" "Green"
}

function Write-Warning {
    param([string]$Message)
    Write-ColorOutput "⚠️  $Message" "Yellow"
}

function Write-Error {
    param([string]$Message)
    Write-ColorOutput "❌ $Message" "Red"
}

function Write-Info {
    param([string]$Message)
    Write-ColorOutput "ℹ️  $Message" "Blue"
}

# Check if running as administrator
function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

# Ensure administrator privileges
function Assert-Administrator {
    if (-not (Test-Administrator)) {
        Write-Error "This operation requires administrator privileges."
        Write-Info "Please run PowerShell as Administrator and try again."
        exit 1
    }
}

# Create necessary directories
function New-ServiceDirectories {
    Write-Log "Creating necessary directories..."
    
    @($InstallPath, $ConfigPath, $LogPath) | ForEach-Object {
        if (-not (Test-Path $_)) {
            if ($WhatIf) {
                Write-Info "Would create directory: $_"
            } else {
                New-Item -Path $_ -ItemType Directory -Force | Out-Null
                Write-Log "Created directory: $_"
            }
        } else {
            Write-Log "Directory already exists: $_"
        }
    }
}

# Check if service exists
function Test-ServiceExists {
    return (Get-Service -Name $ServiceName -ErrorAction SilentlyContinue) -ne $null
}

# Check if service executable exists
function Test-ServiceExecutable {
    return Test-Path $ServicePath
}

# Install Windows service
function Install-PrismService {
    Write-Log "Installing Prism Windows service..."
    Assert-Administrator
    
    if (Test-ServiceExists) {
        if ($Force) {
            Write-Warning "Service already exists. Forcing reinstallation..."
            Uninstall-PrismService
        } else {
            Write-Warning "Service already installed. Use -Force to reinstall."
            return
        }
    }
    
    if (-not (Test-ServiceExecutable)) {
        Write-Error "Service executable not found at: $ServicePath"
        Write-Info "Please ensure Prism is properly installed."
        return
    }
    
    New-ServiceDirectories
    
    if ($WhatIf) {
        Write-Info "Would install service: $DisplayName"
        Write-Info "  Service Name: $ServiceName"
        Write-Info "  Executable: $ServicePath"
        Write-Info "  Startup Type: Automatic"
        return
    }
    
    try {
        # Create the service
        New-Service -Name $ServiceName -DisplayName $DisplayName -Description $Description -BinaryPathName "`"$ServicePath`"" -StartupType Automatic
        
        # Configure service recovery options
        & sc.exe failure $ServiceName reset= 30 actions= restart/5000/restart/5000/restart/5000
        
        Write-Success "Prism service installed successfully"
        Write-Info "Service will start automatically on system boot"
        
        # Start the service
        Start-PrismService
    }
    catch {
        Write-Error "Failed to install service: $_"
    }
}

# Uninstall Windows service
function Uninstall-PrismService {
    Write-Log "Uninstalling Prism Windows service..."
    Assert-Administrator
    
    if (-not (Test-ServiceExists)) {
        Write-Warning "Service not installed"
        return
    }
    
    if ($WhatIf) {
        Write-Info "Would uninstall service: $DisplayName"
        return
    }
    
    try {
        # Stop service if running
        $service = Get-Service -Name $ServiceName
        if ($service.Status -eq "Running") {
            Write-Log "Stopping service..."
            Stop-Service -Name $ServiceName -Force
        }
        
        # Remove service
        Remove-Service -Name $ServiceName
        Write-Success "Prism service uninstalled successfully"
    }
    catch {
        Write-Error "Failed to uninstall service: $_"
    }
}

# Start service
function Start-PrismService {
    Write-Log "Starting Prism service..."
    
    if (-not (Test-ServiceExists)) {
        Write-Error "Service not installed. Run 'install' first."
        return
    }
    
    if ($WhatIf) {
        Write-Info "Would start service: $DisplayName"
        return
    }
    
    try {
        Start-Service -Name $ServiceName
        Write-Success "Prism service started successfully"
    }
    catch {
        Write-Error "Failed to start service: $_"
        Write-Info "Check the Windows Event Log for more details."
    }
}

# Stop service
function Stop-PrismService {
    Write-Log "Stopping Prism service..."
    
    if (-not (Test-ServiceExists)) {
        Write-Warning "Service not installed"
        return
    }
    
    if ($WhatIf) {
        Write-Info "Would stop service: $DisplayName"
        return
    }
    
    try {
        Stop-Service -Name $ServiceName -Force
        Write-Success "Prism service stopped successfully"
    }
    catch {
        Write-Error "Failed to stop service: $_"
    }
}

# Restart service
function Restart-PrismService {
    Write-Log "Restarting Prism service..."
    
    Stop-PrismService
    Start-Sleep -Seconds 3
    Start-PrismService
}

# Get service status
function Get-PrismServiceStatus {
    Write-Log "Prism Service Status:"
    Write-Host ""
    
    if (Test-ServiceExists) {
        $service = Get-Service -Name $ServiceName
        
        Write-Success "📦 Service: Installed"
        Write-Host "   Service Name: $($service.Name)"
        Write-Host "   Display Name: $($service.DisplayName)"
        Write-Host "   Status: $($service.Status)"
        Write-Host "   Start Type: $((Get-WmiObject -Class Win32_Service -Filter "Name='$ServiceName'").StartMode)"
        Write-Host "   Executable: $ServicePath"
        Write-Host "   Config Path: $ConfigPath"
        Write-Host "   Log Path: $LogPath"
        Write-Host ""
        
        # Show process information if running
        if ($service.Status -eq "Running") {
            $process = Get-WmiObject -Class Win32_Service -Filter "Name='$ServiceName'"
            if ($process.ProcessId) {
                Write-ColorOutput "🟢 Process ID: $($process.ProcessId)" "Green"
            }
        }
        
        # Show recent Windows Event Log entries
        $logEntries = Get-EventLog -LogName Application -Source $ServiceName -Newest 5 -ErrorAction SilentlyContinue
        if ($logEntries) {
            Write-Info "📝 Recent Log Entries:"
            $logEntries | ForEach-Object {
                $levelIcon = switch ($_.EntryType) {
                    "Error" { "🔴" }
                    "Warning" { "🟡" }
                    "Information" { "🟢" }
                    default { "ℹ️" }
                }
                Write-Host "   $levelIcon $($_.TimeGenerated) - $($_.Message.Substring(0, [Math]::Min($_.Message.Length, 100)))..."
            }
        }
    } else {
        Write-Error "❌ Service: Not installed"
    }
    
    Write-Host ""
}

# Show service logs
function Show-PrismServiceLogs {
    Write-Log "Showing Prism service logs..."
    
    # Check Windows Event Log
    try {
        $logEntries = Get-EventLog -LogName Application -Source $ServiceName -Newest 50 -ErrorAction Stop
        
        Write-Info "📝 Windows Event Log Entries:"
        $logEntries | ForEach-Object {
            $levelColor = switch ($_.EntryType) {
                "Error" { "Red" }
                "Warning" { "Yellow" }
                "Information" { "Green" }
                default { "White" }
            }
            
            Write-ColorOutput "[$($_.TimeGenerated)] $($_.EntryType): $($_.Message)" $levelColor
        }
    }
    catch {
        Write-Warning "No event log entries found for Prism service"
    }
    
    # Check daemon log files
    $logFiles = Get-ChildItem -Path $LogPath -Filter "*.log" -ErrorAction SilentlyContinue
    if ($logFiles) {
        Write-Host ""
        Write-Info "📁 Daemon Log Files:"
        
        $logFiles | ForEach-Object {
            Write-Host ""
            Write-ColorOutput "--- $($_.Name) ---" "Cyan"
            Get-Content $_.FullName -Tail 20 | Write-Host
        }
    } else {
        Write-Warning "No daemon log files found in $LogPath"
    }
}

# Validate service configuration
function Test-PrismServiceConfiguration {
    Write-Log "Validating Prism service configuration..."
    Write-Host ""
    
    $errors = 0
    
    # Check if running as administrator
    if (Test-Administrator) {
        Write-Success "✅ Administrator privileges: Available"
    } else {
        Write-Warning "⚠️  Administrator privileges: Not available (required for service management)"
    }
    
    # Check service executable
    if (Test-ServiceExecutable) {
        Write-Success "✅ Service executable: Found at $ServicePath"
        
        # Check file version
        try {
            $version = (Get-Item $ServicePath).VersionInfo.FileVersion
            Write-Host "   Version: $version"
        }
        catch {
            Write-Warning "   Could not retrieve version information"
        }
    } else {
        Write-Error "❌ Service executable: Not found at $ServicePath"
        $errors++
    }
    
    # Check daemon executable  
    $daemonPath = Join-Path $InstallPath "cwsd.exe"
    if (Test-Path $daemonPath) {
        Write-Success "✅ Daemon executable: Found at $daemonPath"
    } else {
        Write-Error "❌ Daemon executable: Not found at $daemonPath"
        $errors++
    }
    
    # Check directories
    @(
        @{Path = $InstallPath; Name = "Install directory"}
        @{Path = $ConfigPath; Name = "Config directory"}
        @{Path = $LogPath; Name = "Log directory"}
    ) | ForEach-Object {
        if (Test-Path $_.Path) {
            Write-Success "✅ $($_.Name): $($_.Path) exists"
        } else {
            Write-Warning "⚠️  $($_.Name): $($_.Path) does not exist (will be created if needed)"
        }
    }
    
    # Check service registration
    if (Test-ServiceExists) {
        Write-Success "✅ Service registration: Installed"
        
        $service = Get-WmiObject -Class Win32_Service -Filter "Name='$ServiceName'"
        if ($service.StartMode -eq "Auto") {
            Write-Success "   Auto-start: Enabled"
        } else {
            Write-Warning "   Auto-start: Disabled (StartMode: $($service.StartMode))"
        }
    } else {
        Write-Error "❌ Service registration: Not installed"
        $errors++
    }
    
    Write-Host ""
    if ($errors -eq 0) {
        Write-Success "🎉 Service configuration is valid!"
    } else {
        Write-Error "❌ Found $errors configuration errors"
        return $false
    }
    
    return $true
}

# Show help
function Show-Help {
    Write-Host @"
Prism Windows Service Manager

USAGE:
    windows-service-manager.ps1 <command> [options]

COMMANDS:
    install     Install and start Prism service (requires admin)
    uninstall   Stop and uninstall Prism service (requires admin)
    start       Start the service (requires admin)
    stop        Stop the service (requires admin) 
    restart     Restart the service (requires admin)
    status      Show service status and configuration
    logs        Show service logs from Windows Event Log and daemon files
    validate    Validate service configuration
    help        Show this help message

OPTIONS:
    -Force      Force operation (e.g., reinstall existing service)
    -Verbose    Show verbose output
    -WhatIf     Show what would be done without actually doing it

EXAMPLES:
    # Install service (requires admin PowerShell)
    .\windows-service-manager.ps1 install
    
    # Check service status
    .\windows-service-manager.ps1 status
    
    # View service logs
    .\windows-service-manager.ps1 logs
    
    # Validate configuration
    .\windows-service-manager.ps1 validate

NOTES:
    - Service runs as Local System account
    - Service starts automatically on system boot
    - Service automatically restarts if daemon crashes
    - Logs are written to Windows Event Log and daemon log files
    - Configuration is stored in %ProgramData%\Prism\
"@
}

# Main command dispatcher
switch ($Command.ToLower()) {
    "install" {
        Install-PrismService
    }
    "uninstall" {
        Uninstall-PrismService
    }
    "start" {
        Start-PrismService
    }
    "stop" {
        Stop-PrismService
    }
    "restart" {
        Restart-PrismService
    }
    "status" {
        Get-PrismServiceStatus
    }
    "logs" {
        Show-PrismServiceLogs
    }
    "validate" {
        Test-PrismServiceConfiguration
    }
    "help" {
        Show-Help
    }
    default {
        Write-Error "Unknown command: $Command"
        Write-Host ""
        Show-Help
        exit 1
    }
}

exit 0