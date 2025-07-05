#!/bin/bash

# VictoriaMetrics Installation and Service Setup Script
set -euo pipefail

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SERVICE_FILE="$PROJECT_DIR/deployment/victoria-metrics/victoria-metrics.service"
SYSTEMD_DIR="/etc/systemd/system"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() { echo -e "${GREEN}[INFO]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[WARN]${NC} $1"; }
print_error() { echo -e "${RED}[ERROR]${NC} $1"; }
print_step() { echo -e "${BLUE}[STEP]${NC} $1"; }


# Check if running as root for system operations
check_sudo() {
    if [[ $EUID -eq 0 ]]; then
        print_error "Don't run this script as root. It will ask for sudo when needed."
        exit 1
    fi
}

# Create user
create_users() {
    print_step "Creating system users..."
    
    # VictoriaMetrics user
    if id "victoria-metrics" &>/dev/null; then
        print_warning "User victoria-metrics already exists"
    else
        if sudo useradd --system --shell /bin/false --home-dir /var/lib/victoria-metrics victoria-metrics; then
            print_status "User victoria-metrics created"
        else
            print_error "Failed to create victoria-metrics user"
            return 1
        fi
    fi

    # System monitor user
    if id "system-monitor" &>/dev/null; then
        print_warning "User system-monitor already exists"
    else
        if sudo useradd --system --shell /bin/false --home-dir /var/lib/system-monitor system-monitor; then
            print_status "User system-monitor created"
        else
            print_error "Failed to create system-monitor user"
            return 1
        fi
    fi
    
    # Give system-monitor access to your project directory
    print_status "Setting up project directory permissions..."
    
    # Add system-monitor to aurel group so it can read the project for now 
    # in the future we should build it and copy the build file to /etc/
    if sudo usermod -a -G aurel system-monitor; then
        print_status "Added system-monitor to aurel group"
    else
        print_warning "Failed to add system-monitor to aurel group, continuing..."
    fi
    
    # Ensure group read access to project directory
    if chmod g+rx "$PROJECT_DIR" 2>/dev/null; then
        print_status "Set group permissions on project directory"
    else
        print_warning "Could not set group permissions on project directory"
    fi
    
    # These operations are best-effort, so we use || true
    find "$PROJECT_DIR" -type d -exec chmod g+rx {} \; 2>/dev/null || true
    find "$PROJECT_DIR" -type f -exec chmod g+r {} \; 2>/dev/null || true
    
    # Make sure go.mod and go.sum are readable
    chmod g+r "$PROJECT_DIR/go.mod" "$PROJECT_DIR/go.sum" 2>/dev/null || true
    
    print_status "Permissions configured for system-monitor user" 
}

# Create data directory
create_directories() {
    print_step "Creating directories..."
    
    # VictoriaMetrics directories
    sudo mkdir -p /var/lib/victoria-metrics
    sudo chown victoria-metrics:victoria-metrics /var/lib/victoria-metrics
    sudo chmod 755 /var/lib/victoria-metrics
    
    # System monitor directories
    sudo mkdir -p /opt/system-monitor/{bin,config}
    sudo mkdir -p /var/log/system-monitor
    sudo chown -R system-monitor:system-monitor /opt/system-monitor
    sudo chown system-monitor:system-monitor /var/log/system-monitor
    sudo chmod 755 /opt/system-monitor /var/log/system-monitor
    
    # Config directories
    sudo mkdir -p /etc/victoria-metrics
    sudo mkdir -p /etc/system-monitor
    
    print_status "Directories created"
}

# Download and install binary
download_victoria_metrics() {
    print_status "Downloading VictoriaMetrics binary..."
    
    # Create temp directory
    TEMP_DIR=$(mktemp -d)
    cd "$TEMP_DIR"
    
    # Download latest ARM64 binary
    LATEST_URL="https://github.com/VictoriaMetrics/VictoriaMetrics/releases/download/v1.119.0/victoria-metrics-linux-arm64-v1.119.0.tar.gz"
    
    if wget -q "$LATEST_URL"; then
        tar -xzf victoria-metrics-linux-arm64-v1.119.0.tar.gz
        sudo cp victoria-metrics-prod /usr/local/bin/
        sudo chown root:root /usr/local/bin/victoria-metrics-prod
        sudo chmod 755 /usr/local/bin/victoria-metrics-prod
        print_status "Binary installed to /usr/local/bin/victoria-metrics-prod"
    else
        print_error "Failed to download VictoriaMetrics binary"
        exit 1
    fi
    
    # Cleanup
    cd - > /dev/null
    rm -rf "$TEMP_DIR"
}

install_configs() {
    print_step "Installing configurations..."
    
    # VictoriaMetrics config
    if [[ -f "$PROJECT_DIR/deployment/victoria-metrics/victoria-metrics.conf" ]]; then
        sudo mkdir -p /etc/victoria-metrics
        sudo cp "$PROJECT_DIR/deployment/victoria-metrics/victoria-metrics.conf" /etc/victoria-metrics/
        sudo chown root:victoria-metrics /etc/victoria-metrics/victoria-metrics.conf
        sudo chmod 640 /etc/victoria-metrics/victoria-metrics.conf
        print_status "VictoriaMetrics configuration installed"
    fi
    
    # System monitor config
    if [[ -f "$PROJECT_DIR/deployment/system-monitor/system-monitor.conf" ]]; then
        sudo mkdir -p /etc/system-monitor
        sudo cp "$PROJECT_DIR/deployment/system-monitor/system-monitor.conf" /etc/system-monitor/
        sudo chown root:system-monitor /etc/system-monitor/system-monitor.conf
        sudo chmod 640 /etc/system-monitor/system-monitor.conf
        print_status "System monitor configuration installed"
    fi
}
# Install systemd service
install_victoria_metrics() {
    print_status "Installing systemd victoriametrics service..."
    
    if [[ ! -f "$SERVICE_FILE" ]]; then
        print_error "Service file not found: $SERVICE_FILE"
        exit 1
    fi
    
    # Remove existing symlink if it exists
    if [[ -L "$SYSTEMD_DIR/victoria-metrics.service" ]]; then
        print_warning "Removing existing symlink"
        sudo rm "$SYSTEMD_DIR/victoria-metrics.service"
    fi
    
    # Create symlink
    sudo ln -sf "$SERVICE_FILE" "$SYSTEMD_DIR/victoria-metrics.service"
    print_status "Service file symlinked: $SERVICE_FILE -> $SYSTEMD_DIR/victoria-metrics.service"
    
    # Reload systemd
    sudo systemctl daemon-reload
    print_status "systemd configuration reloaded"
}

# Enable and start service
start_services() {
    print_status "Enabling and starting VictoriaMetrics service..."
    
    sudo systemctl enable victoria-metrics
    sudo systemctl start victoria-metrics
    
    # Wait a moment for startup
    sleep 2
    
    # Check status
    if sudo systemctl is-active --quiet victoria-metrics; then
        print_status "✅ VictoriaMetrics service is running!"
        
        # Test endpoint
        if curl -s http://localhost:8428/metrics > /dev/null; then
            print_status "✅ VictoriaMetrics is responding on port 8428"
        else
            print_warning "⚠️  VictoriaMetrics service is running but not responding on port 8428"
        fi
    else
        print_error "❌ Failed to start VictoriaMetrics service"
        print_error "Check logs with: sudo journalctl -u victoria-metrics -n 20"
        exit 1
    fi

    # Start system monitor
    sudo systemctl enable system-monitor
    sudo systemctl start system-monitor
    
    sleep 2
    
    if sudo systemctl is-active --quiet system-monitor; then
        print_status "✅ System monitor is running"
    else
        print_error "❌ System monitor failed to start"
        sudo journalctl -u system-monitor -n 10
        exit 1
    fi
}

verify_installation() {
    print_step "Verifying installation..."
    
    if curl -s http://localhost:8428/metrics > /dev/null; then
        print_status "✅ VictoriaMetrics is responding"
    else
        print_warning "⚠️  VictoriaMetrics is not responding on port 8428"
    fi
    
    echo ""
    print_status "Service Status:"
    sudo systemctl status victoriametrics system-monitor --no-pager -l
}

# Main installation flow
main() {
    print_status "Starting VictoriaMetrics installation from project directory: $PROJECT_DIR"
    
    check_sudo
    create_users
    create_directories
    download_victoria_metrics
    install_victoria_metrics
    install_configs
    start_services
    verify_installation
    
    print_status "🎉 VictoriaMetrics installation completed!"
    echo ""
    echo "Useful commands:"
    echo "  sudo systemctl status victoria-metrics    # Check status"
    echo "  sudo journalctl -u victoria-metrics -f    # View logs"
    echo "  curl http://localhost:8428/metrics       # Test endpoint"
    echo "  sudo systemctl restart victoria-metrics   # Restart service"
    echo ""
    echo "Web UI available at: http://$(hostname -I | awk '{print $1}'):8428"
}

# Run main function
main "$@";