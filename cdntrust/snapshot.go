package cdntrust

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v3"
)

// Config configures Load.
type Config struct {
	// Providers whose edge ranges are trusted; at least one.
	Providers []Provider

	// SnapshotFile, when set, keeps the last fully successful fetch on disk.
	// It is rewritten after every fetch in which all providers succeeded. When
	// a provider still fails after its retries, its ranges are read from the
	// snapshot instead — with a warning — so a provider outage does not stop
	// the app from booting. A provider the snapshot does not cover still fails
	// the call: an unknown trust set is never used.
	SnapshotFile string

	// Attempts is how many times each source is tried when failures are
	// transient (network errors, 5xx, 429). Defaults to 3.
	Attempts int

	// HTTPClient performs the requests. Defaults to a client bounded by
	// DefaultTimeout.
	HTTPClient *http.Client

	// Logger receives the warning when snapshot ranges are used. Defaults to
	// slog.Default(): Load usually runs before GOE's logger exists.
	Logger *slog.Logger
}

// snapshot is the SnapshotFile format.
type snapshot struct {
	FetchedAt time.Time             `json:"fetched_at"`
	Providers map[Provider][]string `json:"providers"`
}

// Load fetches the providers' edge ranges like TrustProxyConfig, with retries
// and, when cfg.SnapshotFile is set, a last-known-good fallback.
func Load(ctx context.Context, cfg Config) (fiber.TrustProxyConfig, error) {
	if err := checkProviders(cfg.Providers); err != nil {
		return fiber.TrustProxyConfig{}, err
	}
	attempts := cfg.Attempts
	if attempts <= 0 {
		attempts = defaultAttempts
	}
	logger := cfg.Logger
	if logger == nil {
		logger = slog.Default()
	}
	ctx, cancel := withDefaultTimeout(ctx)
	defer cancel()

	results, errs := fetchEach(ctx, (&Fetcher{HTTPClient: cfg.HTTPClient}).client(), cfg.Providers, attempts)

	if len(errs) == 0 {
		if cfg.SnapshotFile != "" {
			if err := writeSnapshot(cfg.SnapshotFile, results); err != nil {
				logger.Warn("cdntrust: could not write the snapshot", "file", cfg.SnapshotFile, "error", err)
			}
		}
		return trustConfig(results)
	}

	var snap snapshot
	if cfg.SnapshotFile != "" {
		var err error
		if snap, err = readSnapshot(cfg.SnapshotFile); err != nil && !errors.Is(err, os.ErrNotExist) {
			logger.Warn("cdntrust: could not read the snapshot", "file", cfg.SnapshotFile, "error", err)
		}
	}
	for _, provider := range cfg.Providers {
		fetchErr, failed := errs[provider]
		if !failed {
			continue
		}
		ranges, err := validate(snap.Providers[provider])
		if err != nil || len(ranges) == 0 {
			if cfg.SnapshotFile == "" {
				return fiber.TrustProxyConfig{}, fetchErr
			}
			return fiber.TrustProxyConfig{}, fmt.Errorf("%w (no last-known-good ranges in %s)", fetchErr, cfg.SnapshotFile)
		}
		results[provider] = ranges
		logger.Warn("cdntrust: provider unreachable; using last-known-good ranges",
			"provider", string(provider),
			"error", fetchErr.Error(),
			"fetched_at", snap.FetchedAt.Format(time.RFC3339),
		)
	}
	return trustConfig(results)
}

func trustConfig(results map[Provider][]string) (fiber.TrustProxyConfig, error) {
	ranges, err := union(results)
	if err != nil {
		return fiber.TrustProxyConfig{}, err
	}
	return fiber.TrustProxyConfig{Proxies: ranges}, nil
}

func readSnapshot(file string) (snapshot, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return snapshot{}, err
	}
	var snap snapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return snapshot{}, fmt.Errorf("decoding %s: %w", file, err)
	}
	return snap, nil
}

// writeSnapshot replaces file atomically, so a crash mid-write never leaves a
// truncated snapshot behind.
func writeSnapshot(file string, results map[Provider][]string) error {
	data, err := json.MarshalIndent(snapshot{FetchedAt: time.Now().UTC(), Providers: results}, "", "  ")
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(file), ".cdntrust-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), file)
}
