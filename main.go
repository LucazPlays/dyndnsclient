package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
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

const (
	defaultBinaryURL   = "https://raw.githubusercontent.com/LucazPlays/dyndnsclient/refs/heads/main/dyndns-client-linux"
	defaultInstallPath = "/usr/local/bin/dyndns-client"
)

func main() {
	setupCmd := flag.Bool("setup", false, "Run setup wizard")
	installCmd := flag.Bool("install", false, "Install as systemd service")
	uninstallCmd := flag.Bool("uninstall", false, "Uninstall systemd service")
	updateCmd := flag.Bool("update", false, "Self-update installed binary from GitHub")
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

	if *updateCmd {
		if err := performSelfUpdate(); err != nil {
			log.Fatalf("Update failed: %v", err)
		}
		fmt.Println("Update successful")
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
	vals := url.Values{}

	if config.IPVersion == 4 || config.IPVersion == 46 {
		ipv4, err := getIPv4Address()
		if err != nil {
			log.Printf("Failed to get IPv4: %v", err)
		} else if ipv4 != "" {
			vals.Set("ipv4", ipv4)
			log.Printf("Got IPv4: %s", ipv4)
		}
	}

	if config.IPVersion == 6 || config.IPVersion == 46 {
		ipv6, err := getIPv6Address()
		if err != nil {
			log.Printf("Failed to get IPv6: %v", err)
		} else if ipv6 != "" {
			vals.Set("ipv6", ipv6)
			log.Printf("Got IPv6: %s", ipv6)
		}
	}

	if len(vals) == 0 {
		log.Println("No addresses found")
		return
	}

	oldAddr := loadLastAddress()
	newAddr := vals.Encode()

	if oldAddr == newAddr && oldAddr != "" {
		log.Println("Address unchanged, skipping update")
		return
	}

	u := &url.URL{
		Scheme: "https",
		Host:   "dynv6.com",
		Path:   "/api/update",
	}
	q := u.Query()
	q.Set("hostname", config.Hostname)
	q.Set("token", config.Token)
	for k, vs := range vals {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	u.RawQuery = q.Encode()

	log.Printf("Sending update to DynV6...")
	status, body, err := sendRequest(u.String())
	if err != nil {
		log.Printf("Failed to send update: %v", err)
		return
	}

	log.Printf("DynV6 response (%d): %s", status, body)
	if status < 200 || status >= 300 {
		log.Printf("Update failed with status %d", status)
		return
	}

	saveLastAddress(newAddr)
	log.Println("Update successful")
}

func getIPv4Address() (string, error) {
	// Prefer using an HTTP client instead of calling out to curl
	resp, err := http.Get("https://api.ipify.org")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	ip := strings.TrimSpace(string(b))
	if ip == "" {
		return "", fmt.Errorf("empty IPv4 response")
	}
	// validate basic IP
	if net.ParseIP(ip) == nil {
		return "", fmt.Errorf("invalid IPv4 response: %s", ip)
	}
	return ip, nil
}

func getIPv6Address() (string, error) {
	// Try to read from 'ip' command, fall back to using net.Interfaces
	cmd := exec.Command("ip", "-6", "addr", "list", "scope", "global")
	output, err := cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "inet6") {
				fields := strings.Fields(line)
				if len(fields) >= 2 {
					addr := fields[1]
					// strip prefix length
					if i := strings.Index(addr, "/"); i != -1 {
						addr = addr[:i]
					}
					if ip := net.ParseIP(addr); ip != nil && ip.To16() != nil {
						return addr, nil
					}
				}
			}
		}
	}

	// Fallback: enumerate interfaces
	ifs, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, itf := range ifs {
		addrs, _ := itf.Addrs()
		for _, a := range addrs {
			s := a.String()
			if idx := strings.Index(s, "/"); idx != -1 {
				s = s[:idx]
			}
			ip := net.ParseIP(s)
			if ip != nil && ip.To16() != nil && ip.To4() == nil {
				return s, nil
			}
		}
	}
	return "", fmt.Errorf("no IPv6 address found")
}

func sendRequest(u string) (int, string, error) {
	// Use net/http client
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return 0, "", err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, "", err
	}
	return resp.StatusCode, string(b), nil
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
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	data, err := os.ReadFile(filepath.Join(home, ".dyndns-client.addr"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func saveLastAddress(addr string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	// ignore error when writing cache
	_ = os.WriteFile(filepath.Join(home, ".dyndns-client.addr"), []byte(addr), 0600)
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

func performSelfUpdate() error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("this operation requires root privileges. Use sudo")
	}

	// Download to a temporary file
	tmpFile := "/tmp/dyndns-client.new"
	bakFile := "/usr/local/bin/dyndns-client.bak"
	installPath := defaultInstallPath

	resp, err := http.Get(defaultBinaryURL)
	if err != nil {
		return fmt.Errorf("failed to download binary: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("failed to download binary: status %d", resp.StatusCode)
	}

	f, err := os.Create(tmpFile)
	if err != nil {
		return fmt.Errorf("failed to create tmp file: %v", err)
	}
	defer f.Close()

	n, err := io.Copy(f, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write tmp file: %v", err)
	}
	if n == 0 {
		return fmt.Errorf("downloaded file is empty")
	}

	if err := f.Chmod(0755); err != nil {
		return fmt.Errorf("failed to set executable bit: %v", err)
	}

	// Backup existing binary
	if _, err := os.Stat(installPath); err == nil {
		if err := copyFile(installPath, bakFile); err != nil {
			return fmt.Errorf("failed to backup existing binary: %v", err)
		}
	}

	// Replace binary
	if err := os.Rename(tmpFile, installPath); err != nil {
		// try to restore backup
		_ = os.Rename(bakFile, installPath)
		return fmt.Errorf("failed to replace binary: %v", err)
	}

	if err := os.Chown(installPath, 0, 0); err != nil {
		return fmt.Errorf("failed to chown installed binary: %v", err)
	}

	if err := os.Chmod(installPath, 0755); err != nil {
		return fmt.Errorf("failed to chmod installed binary: %v", err)
	}

	return nil
}

func copyFile(src, dst string) error {
	srcF, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcF.Close()

	dstF, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstF.Close()

	if _, err := io.Copy(dstF, srcF); err != nil {
		return err
	}
	fi, err := srcF.Stat()
	if err == nil {
		_ = dstF.Chmod(fi.Mode())
	}
	return nil
}
