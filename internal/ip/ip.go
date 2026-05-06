package ip

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os/exec"
	"strings"
	"time"
)

func GetIPv4() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.ipify.org")
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
	if net.ParseIP(ip) == nil {
		return "", fmt.Errorf("invalid IPv4 response: %s", ip)
	}
	return ip, nil
}

func GetIPv6() (string, error) {
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
					if i := strings.Index(addr, "/"); i != -1 {
						addr = addr[:i]
					}
					if ip := net.ParseIP(addr); ip != nil && ip.To16() != nil && !ip.IsLoopback() {
						return addr, nil
					}
				}
			}
		}
	}

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
			if ip != nil && ip.To16() != nil && ip.To4() == nil && !ip.IsLoopback() {
				if !strings.HasPrefix(s, "fe80:") {
					return s, nil
				}
			}
		}
	}
	return "", fmt.Errorf("no IPv6 address found")
}
