package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func runSetup() error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("=== DynDNS Client Setup ===")
	fmt.Println()

	fmt.Print("Enter hostname (e.g., myhost.dynv6.net): ")
	hostname, _ := reader.ReadString('\n')
	hostname = strings.TrimSpace(hostname)
	if hostname == "" {
		return fmt.Errorf("hostname is required")
	}

	fmt.Print("Enter API token: ")
	token, _ := reader.ReadString('\n')
	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("API token is required")
	}

	fmt.Println()
	fmt.Println("Select IP version:")
	fmt.Println("1. IPv4 only")
	fmt.Println("2. IPv6 only")
	fmt.Println("3. Both IPv4 and IPv6")
	fmt.Print("Enter choice (1-3): ")
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	var ipVersion int
	switch choice {
	case "1":
		ipVersion = 4
	case "2":
		ipVersion = 6
	case "3":
		ipVersion = 46
	default:
		ipVersion = 46
	}

	fmt.Print("Enter update interval in seconds (default 300): ")
	intervalStr, _ := reader.ReadString('\n')
	intervalStr = strings.TrimSpace(intervalStr)
	interval := 300
	if intervalStr != "" {
		fmt.Sscanf(intervalStr, "%d", &interval)
	}

	config := fmt.Sprintf("hostname=%s\ntoken=%s\nip_version=%d\ninterval=%d\n",
		hostname, token, ipVersion, interval)

	// Save config with restricted permissions
	if err := os.WriteFile("/etc/dyndns-client.conf", []byte(config), 0600); err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}

	fmt.Println()
	fmt.Println("Configuration saved to /etc/dyndns-client.conf")
	fmt.Println()
	fmt.Println("Do you want to install as a systemd service? (y/n)")
	installChoice, _ := reader.ReadString('\n')
	installChoice = strings.TrimSpace(installChoice)

	if strings.ToLower(installChoice) == "y" {
		if err := installService(); err != nil {
			return fmt.Errorf("failed to install service: %v", err)
		}
		fmt.Println("Service installed and started successfully!")
	}

	return nil
}

func installService() error {
	// Check if running as root
	if os.Getuid() != 0 {
		return fmt.Errorf("this operation requires root privileges. Use sudo.")
	}

	// Get the binary path
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}
	exePath, _ = filepath.Abs(exePath)

	// Create systemd service directory
	os.MkdirAll("/etc/systemd/system", 0755)

	// Create service file
	serviceContent := fmt.Sprintf(`[Unit]
Description=DynDNS Client
After=network-online.target
Wants=network-online.target

[Service]
Type=simple
ExecStart=%s
Restart=always
RestartSec=10
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
`, exePath)

	if err := os.WriteFile("/etc/systemd/system/dyndns-client.service", []byte(serviceContent), 0644); err != nil {
		return fmt.Errorf("failed to write service file: %v", err)
	}

	// Reload systemd
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %v", err)
	}

	// Enable and start service
	if err := exec.Command("systemctl", "enable", "dyndns-client").Run(); err != nil {
		return fmt.Errorf("failed to enable service: %v", err)
	}

	if err := exec.Command("systemctl", "start", "dyndns-client").Run(); err != nil {
		return fmt.Errorf("failed to start service: %v", err)
	}

	return nil
}

func uninstallService() error {
	// Check if running as root
	if os.Getuid() != 0 {
		return fmt.Errorf("this operation requires root privileges. Use sudo.")
	}

	// Stop service
	exec.Command("systemctl", "stop", "dyndns-client").Run()

	// Disable service
	exec.Command("systemctl", "disable", "dyndns-client").Run()

	// Remove service file
	os.Remove("/etc/systemd/system/dyndns-client.service")

	// Reload systemd
	exec.Command("systemctl", "daemon-reload").Run()

	// Remove config
	os.Remove("/etc/dyndns-client.conf")

	fmt.Println("DynDNS client uninstalled successfully")
	return nil
}
