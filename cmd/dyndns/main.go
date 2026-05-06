package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	"dyndns-client/internal/daemon"
	"dyndns-client/internal/system"
	"dyndns-client/internal/updater"
)

func main() {
	setupCmd := flag.Bool("setup", false, "Run setup wizard")
	installCmd := flag.Bool("install", false, "Install as systemd service")
	uninstallCmd := flag.Bool("uninstall", false, "Uninstall systemd service")
	updateCmd := flag.Bool("update", false, "Self-update installed binary from GitHub")
	serviceCmd := flag.String("service", "", "Service action: start, stop, restart, status")
	flag.Parse()

	if *setupCmd {
		if err := system.RunSetup(); err != nil {
			slog.Error("Setup failed", "err", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *installCmd {
		if err := system.InstallService(); err != nil {
			slog.Error("Install failed", "err", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *uninstallCmd {
		if err := system.UninstallService(); err != nil {
			slog.Error("Uninstall failed", "err", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if *updateCmd {
		updated, err := updater.PerformSelfUpdate()
		if err != nil {
			slog.Error("Update failed", "err", err)
			os.Exit(1)
		}
		if updated {
			fmt.Println("Update successful (new version installed).")
		} else {
			fmt.Println("Already up to date.")
		}
		os.Exit(0)
	}

	if *serviceCmd != "" {
		if err := system.ServiceAction(*serviceCmd); err != nil {
			slog.Error("Service action failed", "err", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// Default behavior: start daemon
	daemon.Run()
}
