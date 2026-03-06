#!/bin/bash

# Prism macOS DMG Post-Install Script
# Executed automatically after DMG installation to set up service auto-startup

set -euo pipefail

# Colors for output
red() { echo -e "\033[31m$*\033[0m"; }
green() { echo -e "\033[32m$*\033[0m"; }
yellow() { echo -e "\033[33m$*\033[0m"; }
blue() { echo -e "\033[34m$*\033[0m"; }

# Logging
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $*"
}

# Installation paths
INSTALL_PREFIX="/usr/local"
DAEMON_PATH="$INSTALL_PREFIX/bin/prismd"
CLI_PATH="$INSTALL_PREFIX/bin/cws"
SERVICE_MANAGER="$INSTALL_PREFIX/share/prism/macos-service-manager.sh"

# Check if binaries were installed
check_installation() {
    log "Verifying Prism installation..."
    
    local missing=0
    
    if [[ ! -x "$DAEMON_PATH" ]]; then
        red "❌ Daemon not found at $DAEMON_PATH"
        ((missing++))
    else
        green "✅ Daemon found: $DAEMON_PATH"
    fi
    
    if [[ ! -x "$CLI_PATH" ]]; then
        red "❌ CLI not found at $CLI_PATH"
        ((missing++))
    else
        green "✅ CLI found: $CLI_PATH"
    fi
    
    if [[ $missing -gt 0 ]]; then
        red "Installation verification failed. Service setup skipped."
        return 1
    fi
    
    return 0
}

# Create necessary directories
create_directories() {
    log "Creating necessary directories..."
    
    # User configuration directory
    mkdir -p "$HOME/.prism"
    
    # User log directory
    mkdir -p "$HOME/Library/Logs/prism"
    
    green "✅ Created user directories"
}

# Setup service auto-startup
setup_service() {
    log "Setting up Prism service for auto-startup..."
    
    if [[ -x "$SERVICE_MANAGER" ]]; then
        log "Using service manager: $SERVICE_MANAGER"
        
        # Install user-mode service
        if "$SERVICE_MANAGER" install; then
            green "✅ Prism service installed successfully"
            log "Service will start automatically when you log in"
        else
            yellow "⚠️  Service installation failed or user declined"
            log "You can install the service manually later with:"
            log "  $SERVICE_MANAGER install"
        fi
    else
        yellow "⚠️  Service manager not found at $SERVICE_MANAGER"
        log "Service auto-startup not configured"
        log "You can manually configure service startup if needed"
    fi
}

# Show completion message
show_completion() {
    echo
    green "🎉 Prism installation completed!"
    echo
    blue "📦 What's installed:"
    echo "  • Prism CLI: $CLI_PATH"
    echo "  • Prism Daemon: $DAEMON_PATH"
    echo "  • Service Manager: $SERVICE_MANAGER"
    echo "  • Configuration: $HOME/.prism/"
    echo "  • Logs: $HOME/Library/Logs/prism/"
    echo
    blue "🚀 Getting Started:"
    echo "  cws --help                    # Show CLI help"
    echo "  cws daemon status             # Check daemon status"
    echo
    blue "🔧 Service Management:"
    echo "  $SERVICE_MANAGER status       # Check service status"
    echo "  $SERVICE_MANAGER start        # Start service manually"
    echo "  $SERVICE_MANAGER logs         # View service logs"
    echo
    green "The daemon service has been configured to start automatically!"
}

# Main installation flow
main() {
    log "Prism DMG Post-Install Setup"
    log "Running as user: $(whoami)"
    echo
    
    # Check if installation was successful
    if ! check_installation; then
        exit 1
    fi
    
    # Create user directories
    create_directories
    
    # Setup service auto-startup
    setup_service
    
    # Show completion message
    show_completion
}

# Error handler
error_handler() {
    red "Post-install setup encountered an error on line $1"
    echo
    yellow "Prism has been installed, but service auto-startup may not be configured."
    echo "You can manually configure the service with:"
    echo "  $SERVICE_MANAGER install"
    exit 1
}

# Set error handler
trap 'error_handler $LINENO' ERR

# Run main function
main "$@"