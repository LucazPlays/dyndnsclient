package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Hostname  string
	Token     string
	IPVersion int
	Interval  int
}

const ConfigPath = "/etc/dyndns-client.conf"

func Load() (*Config, error) {
	data, err := os.ReadFile(ConfigPath)
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

func Save(config *Config) error {
	content := fmt.Sprintf("hostname=%s\ntoken=%s\nip_version=%d\ninterval=%d\n",
		config.Hostname, config.Token, config.IPVersion, config.Interval)
	return os.WriteFile(ConfigPath, []byte(content), 0600)
}
