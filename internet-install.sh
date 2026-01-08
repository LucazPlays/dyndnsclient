#!/bin/bash

set -e

# Download URLs - Update these with your actual hosted file URLs
BINARY_URL="https://raw.githubusercontent.com/LucazPlays/dyndnsclient/refs/heads/main/dyndns-client-linux"
INSTALL_SCRIPT_URL="https://raw.githubusercontent.com/LucazPlays/dyndnsclient/refs/heads/main/install.sh"

BINARY_NAME="dyndns-client"
INSTALL_SCRIPT="install.sh"

echo "==================================="
echo "DynDNS Client Internet Installation"
echo "==================================="
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run as root"
    echo "Usage: sudo $0"
    exit 1
fi

echo "Downloading binary from $BINARY_URL..."
wget -q -O "$BINARY_NAME" "$BINARY_URL"
chmod +x "$BINARY_NAME"

echo "Downloading install script from $INSTALL_SCRIPT_URL..."
wget -q -O "$INSTALL_SCRIPT" "$INSTALL_SCRIPT_URL"
chmod +x "$INSTALL_SCRIPT"

echo "Running installation script..."
bash "$INSTALL_SCRIPT"

echo ""
echo "Internet installation complete!"