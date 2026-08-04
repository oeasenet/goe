package lock

import (
	"errors"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"

	"go.oease.dev/goe/v2/contract"
)

// Option configures the lock system from Go code.
//
// Options are applied after environment variables, so a value set in code
// always wins over the matching LOCK_* variable. Environment variables still
// supply everything code does not set, which means an application that passes
// no options behaves exactly as it did before.
//
// The naming rule is mechanical: every field of Config is exposed as
// With<FieldName> (the Default* lock settings drop the "Default" only where
// Config itself does — WithKeyPrefix matches LOCK_KEY_PREFIX).
//
// Credentials are the deliberate exception. WithRedisURL and WithRedisURLs
// reject URLs that embed user:pass — secrets belong in the environment, not
// in source. The two compose: the URL in code declares the topology while
// LOCK_REDIS_USERNAME and LOCK_REDIS_PASSWORD from the environment supply the
// credentials, which apply to every connection mode (single, sentinel,
// cluster, redlock).
//
// Options are applied in the order given, so the last write wins. If any
// option returns an error, the module is not created and goe.New aborts
// startup with every collected error, before any Redis connection is
// attempted.
type Option func(*settings) error

// settings is the mutable state options write to. It is deliberately
// unexported: options are the only way in, which is what keeps credential
// fields out of reach of code configuration.
type settings struct {
	cfg *Config

	// connectionSetInCode records that WithRedisURL or WithRedisURLs supplied
	// the connection, which satisfies ValidateConfig's "connection explicitly
	// configured" requirement the same way a LOCK_REDIS_URL variable would.
	connectionSetInCode bool
}

// resolveConfig runs the full pipeline: defaults -> environment -> options.
// It reports every option error rather than stopping at the first, so a
// developer sees all of their mistakes in one startup failure instead of one
// per run.
func resolveConfig(config contract.Config, opts []Option) (*Config, bool, []error) {
	s := &settings{cfg: LoadConfig(config)}

	var errs []error
	for i, opt := range opts {
		if opt == nil {
			continue
		}
		if err := opt(s); err != nil {
			errs = append(errs, fmt.Errorf("lock option %d: %w", i+1, err))
		}
	}

	return s.cfg, s.connectionSetInCode, errs
}

// urlHasCredentials reports whether rawURL embeds userinfo, using the same
// per-scheme conventions the manager parses with (sentinel URLs use a second
// "@" for userinfo because the first separates the master name).
func urlHasCredentials(rawURL string) bool {
	scheme, rest, ok := strings.Cut(rawURL, "://")
	if !ok {
		return false
	}

	switch scheme {
	case schemeRedis, schemeRedisTLS:
		u, err := url.Parse(rawURL)
		return err == nil && u.User != nil
	case schemeSentinel:
		if slashIdx := strings.LastIndex(rest, "/"); slashIdx != -1 {
			rest = rest[:slashIdx]
		}
		return strings.Count(rest, "@") >= 2
	case schemeCluster:
		if slashIdx := strings.LastIndex(rest, "/"); slashIdx != -1 {
			rest = rest[:slashIdx]
		}
		return strings.Contains(rest, "@")
	default:
		return false
	}
}

// checkCodeURL validates a URL supplied through an Option: it must carry a
// supported scheme and must not embed credentials.
func checkCodeURL(option, rawURL string, allowedSchemes ...string) error {
	if rawURL == "" {
		return fmt.Errorf("%s: url must not be empty", option)
	}
	scheme, _, ok := strings.Cut(rawURL, "://")
	if !ok {
		return fmt.Errorf("%s: invalid Redis URL %q: missing scheme separator", option, rawURL)
	}
	if !slices.Contains(allowedSchemes, scheme) {
		return fmt.Errorf("%s: unsupported URL scheme %q (supported: %s)",
			option, scheme, strings.Join(allowedSchemes, ", "))
	}
	if urlHasCredentials(rawURL) {
		return fmt.Errorf("%s: URL embeds credentials; secrets never belong in code — "+
			"set LOCK_REDIS_USERNAME and LOCK_REDIS_PASSWORD in the environment instead", option)
	}
	return nil
}

// WithRedisURL sets the Redis connection URL, replacing LOCK_REDIS_URL. It
// also overrides LOCK_REDIS_URLS when both are present, because code wins
// over the environment. The scheme selects the connection mode:
//
//   - redis://host[:port][/db]                     single instance
//   - rediss://host[:port][/db]                    single instance over TLS
//   - redis-sentinel://master@host[:port],...[/db] Redis Sentinel
//   - redis-cluster://host[:port],...              Redis Cluster
//
// The URL must not embed credentials: set LOCK_REDIS_USERNAME and
// LOCK_REDIS_PASSWORD in the environment and they are applied to whichever
// topology the URL declares.
func WithRedisURL(rawURL string) Option {
	return func(s *settings) error {
		if err := checkCodeURL("WithRedisURL", rawURL,
			schemeRedis, schemeRedisTLS, schemeSentinel, schemeCluster); err != nil {
			return err
		}
		s.cfg.RedisURL = rawURL
		// The manager prefers RedisURLs over RedisURL; an environment URL list
		// must not silently beat a URL chosen in code.
		s.cfg.RedisURLs = nil
		s.connectionSetInCode = true
		return nil
	}
}

// WithRedisURLs sets multiple independent Redis URLs for the Redlock
// algorithm, replacing LOCK_REDIS_URLS. Each URL must be a redis:// or
// rediss:// single instance — sentinel and cluster schemes cannot participate
// in Redlock.
//
// As with WithRedisURL, the URLs must not embed credentials: use
// LOCK_REDIS_USERNAME and LOCK_REDIS_PASSWORD, which apply to every instance.
func WithRedisURLs(urls ...string) Option {
	return func(s *settings) error {
		if len(urls) == 0 {
			return errors.New("WithRedisURLs: at least one URL is required")
		}
		for _, u := range urls {
			if err := checkCodeURL("WithRedisURLs", u, schemeRedis, schemeRedisTLS); err != nil {
				if scheme, _, ok := strings.Cut(u, "://"); ok &&
					(scheme == schemeSentinel || scheme == schemeCluster) {
					return fmt.Errorf("WithRedisURLs: redlock mode requires redis:// or rediss:// URLs, got %s", scheme)
				}
				return err
			}
		}
		s.cfg.RedisURLs = urls
		s.connectionSetInCode = true
		return nil
	}
}

// WithDefaultExpiry sets the default lock expiry time. Replaces
// LOCK_DEFAULT_EXPIRY.
func WithDefaultExpiry(d time.Duration) Option {
	return func(s *settings) error {
		if d <= 0 {
			return fmt.Errorf("WithDefaultExpiry: %v must be positive", d)
		}
		s.cfg.DefaultExpiry = d
		return nil
	}
}

// WithDefaultTries sets the default number of acquisition attempts. Replaces
// LOCK_DEFAULT_TRIES.
func WithDefaultTries(n int) Option {
	return func(s *settings) error {
		if n <= 0 {
			return fmt.Errorf("WithDefaultTries: %d must be positive", n)
		}
		s.cfg.DefaultTries = n
		return nil
	}
}

// WithDefaultRetryDelay sets the default delay between acquisition attempts.
// Replaces LOCK_DEFAULT_RETRY_DELAY.
func WithDefaultRetryDelay(d time.Duration) Option {
	return func(s *settings) error {
		if d <= 0 {
			return fmt.Errorf("WithDefaultRetryDelay: %v must be positive", d)
		}
		s.cfg.DefaultRetryDelay = d
		return nil
	}
}

// WithDefaultDriftFactor sets the clock drift factor used to shrink lock
// validity. Replaces LOCK_DEFAULT_DRIFT_FACTOR. Must be between 0 and 1
// exclusive; the drift allowance is expiry multiplied by this factor.
func WithDefaultDriftFactor(f float64) Option {
	return func(s *settings) error {
		if f <= 0 || f >= 1 {
			return fmt.Errorf("WithDefaultDriftFactor: %v must be between 0 and 1 exclusive", f)
		}
		s.cfg.DefaultDriftFactor = f
		return nil
	}
}

// WithKeyPrefix sets the prefix for all lock keys. Replaces LOCK_KEY_PREFIX.
func WithKeyPrefix(prefix string) Option {
	return func(s *settings) error {
		if prefix == "" {
			return errors.New("WithKeyPrefix: prefix must not be empty")
		}
		s.cfg.DefaultKeyPrefix = prefix
		return nil
	}
}

// WithPoolSize sets the Redis connection pool size. Replaces LOCK_POOL_SIZE.
func WithPoolSize(size int) Option {
	return func(s *settings) error {
		if size <= 0 {
			return fmt.Errorf("WithPoolSize: size %d must be positive", size)
		}
		s.cfg.PoolSize = size
		return nil
	}
}

// WithTLSInsecureSkipVerify disables TLS certificate verification for
// rediss:// connections. Replaces LOCK_TLS_INSECURE_SKIP_VERIFY. Not
// recommended outside development.
func WithTLSInsecureSkipVerify(skip bool) Option {
	return func(s *settings) error { s.cfg.TLSInsecureSkipVerify = skip; return nil }
}
