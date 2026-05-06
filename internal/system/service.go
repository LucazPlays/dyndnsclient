package system

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func InstallService() error {
	if os.Getuid() != 0 {
		return fmt.Errorf("this operation requires root privileges. Use sudo")
	}

	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %v", err)
	}
	exePath, _ = filepath.Abs(exePath)

	os.MkdirAll("/etc/systemd/system", 0755)

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

	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("failed to reload systemd: %v", err)
	}

	if err := exec.Command("systemctl", "enable", "dyndns-client").Run(); err != nil {
		return fmt.Errorf("failed to enable service: %v", err)
	}

	if err := exec.Command("systemctl", "start", "dyndns-client").Run(); err != nil {
		return fmt.Errorf("failed to start service: %v", err)
	}

	return nil
}

func UninstallService() error {
	if os.Getuid() != 0 {
		return fmt.Errorf("this operation requires root privileges. Use sudo")
	}

	exec.Command("systemctl", "stop", "dyndns-client").Run()
	exec.Command("systemctl", "disable", "dyndns-client").Run()

	os.Remove("/etc/systemd/system/dyndns-client.service")
	exec.Command("systemctl", "daemon-reload").Run()
	os.Remove("/etc/dyndns-client.conf")

	return nil
}

func ServiceAction(action string) error {
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
