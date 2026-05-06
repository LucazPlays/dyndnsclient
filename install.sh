#!/bin/bash
set -e

echo "==================================="
echo "DynDNS Client Installation Script"
echo "==================================="
echo ""

if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run as root"
    echo "Usage: curl -sL <url> | sudo bash"
    exit 1
fi

ARCH=$(uname -m)
case $ARCH in
    x86_64|amd64)
        BIN_ARCH="amd64"
        ;;
    aarch64|arm64)
        BIN_ARCH="arm64"
        ;;
    *)
        echo "Error: Unsupported architecture $ARCH"
        exit 1
        ;;
esac

echo "Detected architecture: $BIN_ARCH"

REPO_URL="https://raw.githubusercontent.com/LucazPlays/dyndnsclient/refs/heads/main"
BINARY_URL="$REPO_URL/dyndns-client-linux-$BIN_ARCH"
HASH_URL="$BINARY_URL.sha256"

INSTALL_DIR="/usr/local/bin"
BINARY_PATH="$INSTALL_DIR/dyndns-client"

echo "Downloading hash..."
EXPECTED_HASH=$(curl -sL "$HASH_URL" | awk '{print $1}')
if [ -z "$EXPECTED_HASH" ]; then
    echo "Warning: Could not fetch hash file. Continuing without hash verification..."
fi

echo "Downloading binary..."
TMP_BIN="/tmp/dyndns-client-new"
curl -sL -o "$TMP_BIN" "$BINARY_URL"

if [ -n "$EXPECTED_HASH" ]; then
    echo "Verifying checksum..."
    ACTUAL_HASH=$(sha256sum "$TMP_BIN" | awk '{print $1}')
    if [ "$EXPECTED_HASH" != "$ACTUAL_HASH" ]; then
        echo "Error: Checksum mismatch! Expected $EXPECTED_HASH, got $ACTUAL_HASH"
        rm -f "$TMP_BIN"
        exit 1
    fi
    echo "Checksum verified."
fi

mkdir -p "$INSTALL_DIR"
mv "$TMP_BIN" "$BINARY_PATH"
chmod +x "$BINARY_PATH"
chown root:root "$BINARY_PATH"

echo ""
echo "Installation complete!"
echo "Running setup wizard..."
echo ""

# Reattach stdin to the terminal so the interactive prompt works
if [ -t 1 ]; then
    exec < /dev/tty
fi

"$BINARY_PATH" --setup
