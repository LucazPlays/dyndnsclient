package provider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"dyndns-client/internal/config"
)

type DynV6 struct{}

func (d *DynV6) Update(cfg *config.Config, ipv4, ipv6 string) error {
	vals := url.Values{}
	if ipv4 != "" {
		vals.Set("ipv4", ipv4)
	}
	if ipv6 != "" {
		vals.Set("ipv6", ipv6)
	}

	if len(vals) == 0 {
		return fmt.Errorf("no addresses provided")
	}

	newAddr := vals.Encode()
	oldAddr := LoadLastAddress()

	if oldAddr == newAddr && oldAddr != "" {
		return nil // unchanged
	}

	u := &url.URL{
		Scheme: "https",
		Host:   "dynv6.com",
		Path:   "/api/update",
	}
	q := u.Query()
	q.Set("hostname", cfg.Hostname)
	q.Set("token", cfg.Token)
	for k, vs := range vals {
		for _, v := range vs {
			q.Add(k, v)
		}
	}
	u.RawQuery = q.Encode()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("update failed with status %d: %s", resp.StatusCode, string(b))
	}

	SaveLastAddress(newAddr)
	return nil
}
