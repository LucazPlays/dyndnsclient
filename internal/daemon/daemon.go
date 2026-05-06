package daemon

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dyndns-client/internal/config"
	"dyndns-client/internal/ip"
	"dyndns-client/internal/provider"
	"dyndns-client/internal/updater"
)

func Run() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("Failed to load config", "err", err)
		os.Exit(1)
	}

	slog.Info("DynDNS Client started", "hostname", cfg.Hostname, "interval", cfg.Interval)

	p := &provider.DynV6{}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	updateTicker := time.NewTicker(time.Duration(cfg.Interval) * time.Second)
	defer updateTicker.Stop()

	var autoUpdateChan <-chan time.Time
	if cfg.AutoUpdate {
		autoUpdateTicker := time.NewTicker(24 * time.Hour)
		defer autoUpdateTicker.Stop()
		autoUpdateChan = autoUpdateTicker.C
	} else {
		autoUpdateChan = make(chan time.Time) // Dummy channel
	}

	doUpdate(cfg, p)

	for {
		select {
		case <-sigChan:
			slog.Info("Shutting down daemon...")
			return
		case <-updateTicker.C:
			doUpdate(cfg, p)
		case <-autoUpdateChan:
			slog.Info("Checking for automatic updates...")
			updated, err := updater.PerformSelfUpdate()
			if err != nil {
				slog.Error("Auto-update failed", "err", err)
			} else if updated {
				slog.Info("Client was auto-updated. Restarting service...")
				os.Exit(0)
			}
		}
	}
}

func doUpdate(cfg *config.Config, p provider.Provider) {
	var ipv4, ipv6 string
	var err error

	if cfg.IPVersion == 4 || cfg.IPVersion == 46 {
		ipv4, err = ip.GetIPv4()
		if err != nil {
			slog.Warn("Failed to get IPv4", "err", err)
		} else if ipv4 != "" {
			slog.Info("Got IPv4", "ip", ipv4)
		}
	}

	if cfg.IPVersion == 6 || cfg.IPVersion == 46 {
		ipv6, err = ip.GetIPv6()
		if err != nil {
			slog.Warn("Failed to get IPv6", "err", err)
		} else if ipv6 != "" {
			slog.Info("Got IPv6", "ip", ipv6)
		}
	}

	if ipv4 == "" && ipv6 == "" {
		slog.Warn("No addresses found, skipping update")
		return
	}

	if err := p.Update(cfg, ipv4, ipv6); err != nil {
		slog.Error("Provider update failed", "err", err)
	} else {
		slog.Info("DNS update processed")
	}
}
