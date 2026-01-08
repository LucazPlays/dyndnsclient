#!/bin/bash

set -e

BINARY_NAME="dyndns-client"
INSTALL_DIR="/usr/local/bin"
SERVICE_NAME="dyndns-client"

echo "====================================="
echo "DynDNS Client Uninstallation Script"
echo "====================================="
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run as root"
    echo "Usage: sudo $0"
    exit 1
fi

# Stop and disable service
echo "Stopping and disabling systemd service..."
systemctl stop "$SERVICE_NAME" 2>/dev/null || true
systemctl disable "$SERVICE_NAME" 2>/dev/null || true

# Remove service file
echo "Removing systemd service file..."
rm -f /etc/systemd/system/"$SERVICE_NAME".service
systemctl daemon-reload

# Remove binary
echo "Removing binary..."
rm -f "$INSTALL_DIR/$BINARY_NAME"

# Remove configuration
echo "Removing configuration..."
rm -f /etc/dyndns-client.conf

# Remove cached address file
echo "Removing cached address file..."
rm -f ~/.dyndns-client.addr

echo ""
echo "Uninstallation complete!"
echo ""
echo "Note: You may also want to remove the source directory:"
echo "  rm -rf $(pwd)"
echo ""

exit 0
