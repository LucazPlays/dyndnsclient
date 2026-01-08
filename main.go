package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	Hostname  string
	Token     string
	IPVersion int
	Interval  int
}

func main() {
	setupCmd := flag.Bool("setup", false, "Run setup wizard")
	installCmd := flag.Bool("install", false, "Install as systemd service")
	uninstallCmd := flag.Bool("uninstall", false, "Uninstall systemd service")
	serviceCmd := flag.String("service", "", "Service action: start, stop, restart, status")
	flag.Parse()

	if *setupCmd {
		if err := runSetup(); err != nil {
			log.Fatalf("Setup failed: %v", err)
		}
		os.Exit(0)
	}

	if *installCmd {
		if err := installService(); err != nil {
			log.Fatalf("Install failed: %v", err)
		}
		os.Exit(0)
	}

	if *uninstallCmd {
		if err := uninstallService(); err != nil {
			log.Fatalf("Uninstall failed: %v", err)
		}
		os.Exit(0)
	}

	if *serviceCmd != "" {
		if err := serviceAction(*serviceCmd); err != nil {
			log.Fatalf("Service action failed: %v", err)
		}
		os.Exit(0)
	}

	runDaemon()
}

func runDaemon() {
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Println("DynDNS Client started")
	log.Printf("Config: hostname=%s, ip_version=%d, interval=%d",
		config.Hostname, config.IPVersion, config.Interval)

	for {
		updateDNS(config)
		time.Sleep(time.Duration(config.Interval) * time.Second)
	}
}

func updateDNS(config *Config) {
	var addresses []string

	if config.IPVersion == 4 || config.IPVersion == 46 {
		ipv4, err := getIPv4Address()
		if err != nil {
			log.Printf("Failed to get IPv4: %v", err)
		} else if ipv4 != "" {
			addresses = append(addresses, "ipv4="+ipv4)
			log.Printf("Got IPv4: %s", ipv4)
		}
	}

	if config.IPVersion == 6 || config.IPVersion == 46 {
		ipv6, err := getIPv6Address()
		if err != nil {
			log.Printf("Failed to get IPv6: %v", err)
		} else if ipv6 != "" {
			addresses = append(addresses, "ipv6="+ipv6)
			log.Printf("Got IPv6: %s", ipv6)
		}
	}

	if len(addresses) == 0 {
		log.Println("No addresses found")
		return
	}

	oldAddr := loadLastAddress()
	newAddr := strings.Join(addresses, "&")

	if oldAddr == newAddr && oldAddr != "" {
		log.Println("Address unchanged, skipping update")
		return
	}

	url := fmt.Sprintf("https://dynv6.com/api/update?hostname=%s&token=%s&%s",
		config.Hostname, config.Token, strings.Join(addresses, "&"))

	log.Printf("Sending update to DynV6...")
	resp, err := sendRequest(url)
	if err != nil {
		log.Printf("Failed to send update: %v", err)
		return
	}

	log.Printf("DynV6 response: %s", resp)
	saveLastAddress(newAddr)
	log.Println("Update successful")
}

func getIPv4Address() (string, error) {
	cmd := exec.Command("curl", "-s", "https://api.ipify.org")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	ip := strings.TrimSpace(string(output))
	if ip == "" {
		return "", fmt.Errorf("empty IPv4 response")
	}
	return ip, nil
}

func getIPv6Address() (string, error) {
	cmd := exec.Command("ip", "-6", "addr", "list", "scope", "global")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "inet6") && !strings.HasPrefix(line, "\tf") {
			fields := strings.Fields(line)
			for _, field := range fields {
				if strings.HasPrefix(field, "2") || strings.HasPrefix(field, "1") {
					return strings.TrimSuffix(field, "/128"), nil
				}
			}
		}
	}
	return "", fmt.Errorf("no IPv6 address found")
}

func sendRequest(url string) (string, error) {
	cmd := exec.Command("curl", "-s", "-w", "\n%{http_code}", url)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func loadConfig() (*Config, error) {
	configPath := "/etc/dyndns-client.conf"
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	config := &Config{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "hostname=") {
			config.Hostname = strings.TrimPrefix(line, "hostname=")
		} else if strings.HasPrefix(line, "token=") {
			config.Token = strings.TrimPrefix(line, "token=")
		} else if strings.HasPrefix(line, "ip_version=") {
			fmt.Sscanf(strings.TrimPrefix(line, "ip_version="), "%d", &config.IPVersion)
		} else if strings.HasPrefix(line, "interval=") {
			fmt.Sscanf(strings.TrimPrefix(line, "interval="), "%d", &config.Interval)
		}
	}

	if config.Interval == 0 {
		config.Interval = 300
	}
	if config.IPVersion == 0 {
		config.IPVersion = 46
	}
	return config, nil
}

func loadLastAddress() string {
	home, _ := os.UserHomeDir()
	data, _ := os.ReadFile(filepath.Join(home, ".dyndns-client.addr"))
	return strings.TrimSpace(string(data))
}

func saveLastAddress(addr string) {
	home, _ := os.UserHomeDir()
	os.WriteFile(filepath.Join(home, ".dyndns-client.addr"), []byte(addr), 0644)
}

func serviceAction(action string) error {
	switch action {
	case "start":
		return exec.Command("systemctl", "start", "dyndns-client").Run()
	case "stop":
		return exec.Command("systemctl", "stop", "dyndns-client").Run()
	case "restart":
		return exec.Command("systemctl", "restart", "dyndns-client").Run()
	case "status":
		return exec.Command("systemctl", "status", "dyndns-client").Run()
	default:
		return fmt.Errorf("unknown action: %s", action)
	}
}
