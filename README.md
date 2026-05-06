# DynDNS Client

A robust, professional Dynamic DNS client written in Go for Linux (Debian/Ubuntu/CentOS, etc.) with native systemd support, multi-architecture builds (x86_64 / ARM64), and automatic updates.

**Quick Install & Setup:** Download and run the smart installer in one line:

```bash
curl -sL https://raw.githubusercontent.com/LucazPlays/dyndnsclient/refs/heads/main/install.sh | sudo bash
```

## Features

- **Multi-Architecture**: Automatically detects your system (amd64 / arm64) and installs the correct binary.
- **Smart Auto-Updater**: Background daemon checks for new versions via SHA256 hashes every 24 hours and updates automatically.
- **DynV6.com API**: First-class support for updating DynV6.
- **Dual-Stack IP Support**: Intelligent IPv4 (via API) and IPv6 (via local network interfaces) address detection.
- **systemd Integration**: Easily install as a reliable background service.
- **Interactive Setup Wizard**: Get up and running in seconds without editing config files manually.

## Installation

### Internet Installer (Recommended)

The recommended way to install is via our smart script. It detects your CPU architecture, verifies file checksums, installs the client, and launches the setup wizard automatically:

```bash
curl -sL https://raw.githubusercontent.com/LucazPlays/dyndnsclient/refs/heads/main/install.sh | sudo bash
```

### From Source

If you prefer building from source:

```bash
# Clone or navigate to the project directory
git clone https://github.com/LucazPlays/dyndnsclient.git
cd dyndnsclient

# Build the binary
make build

# Install the binary
sudo make install
```

## Configuration

### Interactive Setup

If you didn't use the automated install script, you can run the setup wizard manually:

```bash
sudo dyndns-client --setup
```

The setup wizard will prompt you for:
- Hostname (e.g., `myhost.dynv6.net`)
- API token from DynV6
- IP version preference (IPv4 only, IPv6 only, or both)
- Update interval in seconds

### Manual Configuration

If you prefer, you can manually create `/etc/dyndns-client.conf`:

```ini
hostname=myhost.dynv6.net
token=your-api-token
ip_version=46  # 4=IPv4 only, 6=IPv6 only, 46=both
interval=300   # Update interval in seconds
```

## Service Management

### Install as Service

```bash
sudo dyndns-client --install
```

### Control Service

```bash
sudo systemctl start dyndns-client
sudo systemctl stop dyndns-client
sudo systemctl restart dyndns-client
sudo systemctl status dyndns-client
```

### Uninstall

```bash
sudo dyndns-client --uninstall
# Or use the uninstall script if you have the source code
sudo ./uninstall.sh
```

## Updates

### Automatic Updates

The `dyndns-client` daemon will check for updates every 24 hours. If it detects a newer version (based on the SHA256 hash on GitHub), it will download it, verify the hash, replace the binary, and gracefully restart the service.

### Manual Self-Update

You can also force a self-update manually at any time:

```bash
sudo dyndns-client --update
```

## Development

To build binaries for both `amd64` and `arm64` along with their hash checksums, run:

```bash
make hashes
```

## Files

- **Binary**: `/usr/local/bin/dyndns-client`
- **Config**: `/etc/dyndns-client.conf`
- **Service**: `/etc/systemd/system/dyndns-client.service`
- **Cache**: `~/.dyndns-client.addr`

## License

MIT
