#!/bin/bash

# SystemMonitor Installation and Service Setup Script
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SERVICE_FILE="$PROJECT_DIR/deployment/system-monitoring/system-monitor.service"
SYSTEMD_DIR="/etc/systemd/system"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color
print_status() { echo -e "${GREEN}[INFO]${NC} $1"; }
print_step() { echo -e "${BLUE}[STEP]${NC} $1"; }

install_system_monitor() {
    print_step "Building and installing system-monitor..."
    
    # Build the application
    if ! go build -o bin/system-monitor cmd/main.go; then
        print_error "Failed to build system monitor"
        exit 1
    fi
    
    # Stop service if running
    if systemctl is-active --quiet system-monitor 2>/dev/null; then
        sudo systemctl stop system-monitor
        print_status "Stopped running service"
    fi


     # Create symlink
    sudo ln -sf "$SERVICE_FILE" "$SYSTEMD_DIR/system-monitor.service"
    print_status "Service file symlinked: $SERVICE_FILE -> $SYSTEMD_DIR/system-monitor.service"
    
    # Reload systemd
    sudo systemctl daemon-reload
    print_status "systemd configuration reloaded"
    
    sudo cp bin/system-monitor /opt/system-monitor/bin/
    sudo chmod 755 /opt/system-monitor/bin/system-monitor
    sudo chown system-monitor:system-monitor /opt/system-monitor/bin/system-monitor
    
    sudo cp deployment/system-monitoring/system-monitor.conf /etc/system-monitor/

    # Enable and start service
    sudo systemctl enable system-monitor
    sudo systemctl restart system-monitor

    print_status "System monitor binary installed"
}


# Main installation flow
main() {
    print_status "Starting SystemMonitor installation from project directory: $PROJECT_DIR"
    
    install_system_monitor
    
    print_status "🎉 SystemMonitor installation completed!"
}

# Run main function
main "$@";