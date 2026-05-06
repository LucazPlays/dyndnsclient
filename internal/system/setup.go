package system

import (
	"fmt"
	"os"
	"strings"

	"dyndns-client/internal/config"
	"golang.org/x/term"
)

type terminalRW struct {
	in  *os.File
	out *os.File
}

func (trw terminalRW) Read(p []byte) (n int, err error) { return trw.in.Read(p) }
func (trw terminalRW) Write(p []byte) (n int, err error) { return trw.out.Write(p) }

func readLine(t *term.Terminal, prompt string) string {
	t.SetPrompt(prompt)
	line, err := t.ReadLine()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(line)
}

func RunSetup() error {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return fmt.Errorf("stdin is not a terminal, please run the setup in an interactive shell")
	}

	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return fmt.Errorf("failed to initialize terminal: %v", err)
	}
	defer term.Restore(fd, oldState)

	t := term.NewTerminal(terminalRW{os.Stdin, os.Stdout}, "")

	t.Write([]byte("=== DynDNS Client Setup ===\r\n\r\n"))

	hostname := readLine(t, "Enter hostname (e.g., myhost.dynv6.net): ")
	if hostname == "" {
		return fmt.Errorf("hostname is required")
	}

	token := readLine(t, "Enter API token: ")
	if token == "" {
		return fmt.Errorf("API token is required")
	}

	t.Write([]byte("\r\nSelect IP version:\r\n"))
	t.Write([]byte("1. IPv4 only\r\n"))
	t.Write([]byte("2. IPv6 only\r\n"))
	t.Write([]byte("3. Both IPv4 and IPv6\r\n"))

	choice := readLine(t, "Enter choice (1-3): ")

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

	intervalStr := readLine(t, "Enter update interval in seconds (default 300): ")
	interval := 300
	if intervalStr != "" {
		fmt.Sscanf(intervalStr, "%d", &interval)
	}

	autoUpdateStr := readLine(t, "Enable automatic updates? (Y/n): ")
	autoUpdateStr = strings.ToLower(autoUpdateStr)
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

	t.Write([]byte("\r\nConfiguration saved to /etc/dyndns-client.conf\r\n\r\n"))

	installChoice := readLine(t, "Do you want to install as a systemd service? (y/n): ")

	// Restore the terminal state early so standard printing works again for systemd output
	term.Restore(fd, oldState)

	if strings.ToLower(installChoice) == "y" {
		if err := InstallService(); err != nil {
			return fmt.Errorf("failed to install service: %v", err)
		}
		fmt.Println("Service installed and started successfully!")
	}

	return nil
}
