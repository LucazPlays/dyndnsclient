package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
)

const (
	repoBaseURL = "https://raw.githubusercontent.com/LucazPlays/dyndnsclient/refs/heads/main/"
	installPath = "/usr/local/bin/dyndns-client"
)

func getBinaryName() string {
	if runtime.GOARCH == "arm64" || runtime.GOARCH == "aarch64" {
		return "dyndns-client-linux-arm64"
	}
	return "dyndns-client-linux-amd64"
}

func getHash(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	fields := strings.Fields(string(b))
	if len(fields) > 0 {
		return fields[0], nil
	}
	return "", fmt.Errorf("empty hash file")
}

func getLocalHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func PerformSelfUpdate() (bool, error) {
	if os.Geteuid() != 0 {
		return false, fmt.Errorf("this operation requires root privileges. Use sudo")
	}

	binName := getBinaryName()
	remoteBinURL := repoBaseURL + binName
	remoteHashURL := remoteBinURL + ".sha256"

	remoteHash, err := getHash(remoteHashURL)
	if err != nil {
		return false, fmt.Errorf("failed to fetch remote hash: %v", err)
	}

	localHash, err := getLocalHash(installPath)
	if err == nil && localHash == remoteHash {
		// Already up to date
		return false, nil
	}

	tmpFile := "/tmp/dyndns-client.new"
	bakFile := installPath + ".bak"

	resp, err := http.Get(remoteBinURL)
	if err != nil {
		return false, fmt.Errorf("failed to download binary: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("failed to download binary: status %d", resp.StatusCode)
	}

	f, err := os.Create(tmpFile)
	if err != nil {
		return false, fmt.Errorf("failed to create tmp file: %v", err)
	}
	
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		return false, fmt.Errorf("failed to write tmp file: %v", err)
	}
	f.Close()

	downloadedHash, err := getLocalHash(tmpFile)
	if err != nil {
		return false, fmt.Errorf("failed to hash downloaded file: %v", err)
	}
	if downloadedHash != remoteHash {
		return false, fmt.Errorf("hash mismatch: expected %s, got %s", remoteHash, downloadedHash)
	}

	if err := os.Chmod(tmpFile, 0755); err != nil {
		return false, err
	}

	if _, err := os.Stat(installPath); err == nil {
		if err := copyFile(installPath, bakFile); err != nil {
			return false, fmt.Errorf("failed to backup existing binary: %v", err)
		}
	}

	if err := os.Rename(tmpFile, installPath); err != nil {
		_ = os.Rename(bakFile, installPath)
		return false, fmt.Errorf("failed to replace binary: %v", err)
	}

	_ = os.Chown(installPath, 0, 0)
	return true, nil
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
