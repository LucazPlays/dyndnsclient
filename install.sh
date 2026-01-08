#!/bin/bash

set -e

BINARY_NAME="dyndns-client"
INSTALL_DIR="/usr/local/bin"
SERVICE_NAME="dyndns-client"

echo "==================================="
echo "DynDNS Client Installation Script"
echo "==================================="
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run as root"
    echo "Usage: sudo $0"
    exit 1
fi

# Detect architecture
ARCH=$(uname -m)
OS=$(uname -s)

echo "Detected system: $OS ($ARCH)"

# Create installation directory
mkdir -p "$INSTALL_DIR"

# Install binary
echo "Installing binary to $INSTALL_DIR/$BINARY_NAME..."
cp "$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
chmod +x "$INSTALL_DIR/$BINARY_NAME"

# Install systemd service (optional, done via --install flag)
echo ""
echo "Installation complete!"
echo ""
echo "Usage:"
echo "  $INSTALL_DIR/$BINARY_NAME --setup    - Configure the client interactively"
echo "  $INSTALL_DIR/$BINARY_NAME --install  - Install as systemd service"
echo "  $INSTALL_DIR/$BINARY_NAME --uninstall - Remove systemd service"
echo "  sudo systemctl start $SERVICE_NAME   - Start the service"
echo "  sudo systemctl stop $SERVICE_NAME    - Stop the service"
echo "  sudo systemctl restart $SERVICE_NAME - Restart the service"
echo ""

exit 0
