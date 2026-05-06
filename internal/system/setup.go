package system

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"dyndns-client/internal/config"
)

func RunSetup() error {
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

	fmt.Print("Enable automatic updates? (Y/n): ")
	autoUpdateStr, _ := reader.ReadString('\n')
	autoUpdateStr = strings.TrimSpace(strings.ToLower(autoUpdateStr))
	autoUpdate := true
	if autoUpdateStr == "n" || autoUpdateStr == "no" {
		autoUpdate = false
	}

	cfg := &config.Config{
		Hostname:   hostname,
		Token:      token,
		IPVersion:  ipVersion,
		Interval:   interval,
		AutoUpdate: autoUpdate,
	}

	if err := config.Save(cfg); err != nil {
		return fmt.Errorf("failed to write config: %v", err)
	}

	fmt.Println("\nConfiguration saved to /etc/dyndns-client.conf\n")
	fmt.Println("Do you want to install as a systemd service? (y/n)")
	installChoice, _ := reader.ReadString('\n')
	installChoice = strings.TrimSpace(installChoice)

	if strings.ToLower(installChoice) == "y" {
		if err := InstallService(); err != nil {
			return fmt.Errorf("failed to install service: %v", err)
		}
		fmt.Println("Service installed and started successfully!")
	}

	return nil
}
