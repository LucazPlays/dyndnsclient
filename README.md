# DynDNS Client

**Quick install:** Download and run the internet installer in one line:

```bash
wget -qO- https://raw.githubusercontent.com/LucazPlays/dyndnsclient/refs/heads/main/internet-install.sh | sudo bash
```

A Dynamic DNS client written in Go for Linux Debian with systemd support.

## Features

- Supports DynV6.com API
- IPv4 and/or IPv6 address detection
- Configurable update interval
- systemd service integration
- Interactive setup wizard

## Installation

### From Source

```bash
# Clone or navigate to the project directory
cd dyndnsclient

# Run the installation script
sudo ./install.sh

# Or use make
make
sudo make install
```

### Manual Installation

```bash
# Build the binary
go build -o dyndns-client .

# Install the binary
sudo cp dyndns-client /usr/local/bin/
sudo chmod +x /usr/local/bin/dyndns-client
```

## Configuration

### Interactive Setup

```bash
sudo /usr/local/bin/dyndns-client --setup
```

The setup wizard will prompt you for:
- Hostname (e.g., `myhost.dynv6.net`)
- API token from DynV6
- IP version preference (IPv4 only, IPv6 only, or both)
- Update interval in seconds

### Manual Configuration

Create `/etc/dyndns-client.conf`:

```ini
hostname=myhost.dynv6.net
token=your-api-token
ip_version=46  # 4=IPv4 only, 6=IPv6 only, 46=both
interval=300   # seconds
```

## Service Management

### Install as Service

```bash
sudo /usr/local/bin/dyndns-client --install
```

### Control Service

```bash
# Start the service
sudo systemctl start dyndns-client

# Stop the service
sudo systemctl stop dyndns-client

# Restart the service
sudo systemctl restart dyndns-client

# Check status
sudo systemctl status dyndns-client
```

### Uninstall

```bash
sudo /usr/local/bin/dyndns-client --uninstall

# Or use the uninstall script
sudo ./uninstall.sh

# Or use make
make uninstall
```

## Usage Without Service

Run in foreground (for testing):

```bash
/usr/local/bin/dyndns-client
```

## Files

- Binary: `/usr/local/bin/dyndns-client`
- Config: `/etc/dyndns-client.conf`
- Service: `/etc/systemd/system/dyndns-client.service`
- Cache: `~/.dyndns-client.addr`

## License

MIT
