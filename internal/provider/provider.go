package provider

import (
	"os"
	"path/filepath"
	"strings"

	"dyndns-client/internal/config"
)

type Provider interface {
	Update(cfg *config.Config, ipv4, ipv6 string) error
}

func GetCachePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/.dyndns-client.addr"
	}
	return filepath.Join(home, ".dyndns-client.addr")
}

func LoadLastAddress() string {
	data, err := os.ReadFile(GetCachePath())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

func SaveLastAddress(addr string) {
	_ = os.WriteFile(GetCachePath(), []byte(addr), 0600)
}
