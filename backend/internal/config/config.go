// Package config loads the service configuration from environment variables.
package config

import (
	"fmt"
	"log/slog"
	"net/netip"
	"strconv"
	"strings"
	"time"
)

// Config holds every runtime setting of the service.
type Config struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
	// RequestTimeout is how long a handler may run before the request is
	// answered with 503. It is always shorter than WriteTimeout.
	RequestTimeout time.Duration
	LogLevel       slog.Level
	// MaxBodyBytes is the largest request body accepted, in bytes.
	MaxBodyBytes int64
	// RateLimitPerSecond and RateLimitBurst configure the per-client token
	// bucket: a client may send RateLimitBurst requests at once, refilled at
	// RateLimitPerSecond requests per second.
	RateLimitPerSecond int64
	RateLimitBurst     int64
	// TrustedProxies are the networks of reverse proxies (such as nginx) whose
	// X-Forwarded-For header is trusted to carry the real client address.
	// Empty means no proxy is trusted and the TCP peer address is used.
	TrustedProxies []netip.Prefix
}

// LookupFunc returns the value of an environment variable and whether it is set.
// os.LookupEnv satisfies it; tests pass a map-backed function instead.
type LookupFunc func(key string) (string, bool)

// Load builds a Config from the given lookup function, applying defaults for
// unset variables. It fails fast on any invalid value so that a misconfigured
// service never starts.
func Load(lookup LookupFunc) (Config, error) {
	cfg := Config{
		Port:               DefaultPort,
		ReadTimeout:        DefaultReadTimeout,
		WriteTimeout:       DefaultWriteTimeout,
		IdleTimeout:        DefaultIdleTimeout,
		ShutdownTimeout:    DefaultShutdownTimeout,
		RequestTimeout:     DefaultRequestTimeout,
		LogLevel:           DefaultLogLevel,
		MaxBodyBytes:       DefaultMaxBodyBytes,
		RateLimitPerSecond: DefaultRateLimitPerSecond,
		RateLimitBurst:     DefaultRateLimitBurst,
	}

	if err := loadPort(lookup, &cfg.Port); err != nil {
		return Config{}, err
	}

	durations := []struct {
		key    string
		target *time.Duration
	}{
		{EnvReadTimeout, &cfg.ReadTimeout},
		{EnvWriteTimeout, &cfg.WriteTimeout},
		{EnvIdleTimeout, &cfg.IdleTimeout},
		{EnvShutdownTimeout, &cfg.ShutdownTimeout},
		{EnvRequestTimeout, &cfg.RequestTimeout},
	}
	for _, d := range durations {
		if err := loadDuration(lookup, d.key, d.target); err != nil {
			return Config{}, err
		}
	}

	if err := loadLogLevel(lookup, &cfg.LogLevel); err != nil {
		return Config{}, err
	}
	positiveInts := []struct {
		key    string
		target *int64
	}{
		{EnvMaxBodyBytes, &cfg.MaxBodyBytes},
		{EnvRateLimitPerSecond, &cfg.RateLimitPerSecond},
		{EnvRateLimitBurst, &cfg.RateLimitBurst},
	}
	for _, p := range positiveInts {
		if err := loadPositiveInt(lookup, p.key, p.target); err != nil {
			return Config{}, err
		}
	}
	if err := loadTrustedProxies(lookup, &cfg.TrustedProxies); err != nil {
		return Config{}, err
	}

	// The server stops writing after WriteTimeout, so a request timeout that is
	// not shorter could never deliver its 503 response.
	if cfg.RequestTimeout >= cfg.WriteTimeout {
		return Config{}, fmt.Errorf("config: %s (%s) must be shorter than %s (%s)",
			EnvRequestTimeout, cfg.RequestTimeout, EnvWriteTimeout, cfg.WriteTimeout)
	}

	return cfg, nil
}

// Addr returns the TCP address the HTTP server listens on, e.g. ":8080".
func (c Config) Addr() string {
	return fmt.Sprintf(":%d", c.Port)
}

func loadPort(lookup LookupFunc, target *int) error {
	raw, ok := lookup(EnvPort)
	if !ok {
		return nil
	}
	port, err := strconv.Atoi(raw)
	if err != nil || port < minPort || port > maxPort {
		return fmt.Errorf("config: %s must be an integer between %d and %d, got %q", EnvPort, minPort, maxPort, raw)
	}
	*target = port
	return nil
}

func loadDuration(lookup LookupFunc, key string, target *time.Duration) error {
	raw, ok := lookup(key)
	if !ok {
		return nil
	}
	d, err := time.ParseDuration(raw)
	if err != nil || d <= 0 {
		return fmt.Errorf("config: %s must be a positive duration such as \"5s\", got %q", key, raw)
	}
	*target = d
	return nil
}

func loadPositiveInt(lookup LookupFunc, key string, target *int64) error {
	raw, ok := lookup(key)
	if !ok {
		return nil
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || value <= 0 {
		return fmt.Errorf("config: %s must be a positive integer, got %q", key, raw)
	}
	*target = value
	return nil
}

// loadTrustedProxies parses a comma-separated list of networks ("172.16.0.0/12")
// or single addresses ("10.0.0.5"); a single address is a network of one host.
func loadTrustedProxies(lookup LookupFunc, target *[]netip.Prefix) error {
	raw, ok := lookup(EnvTrustedProxies)
	if !ok {
		return nil
	}
	var prefixes []netip.Prefix
	for _, item := range strings.Split(raw, trustedProxySeparator) {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		prefix, err := parseNetwork(item)
		if err != nil {
			return fmt.Errorf("config: %s must be a comma-separated list of IP addresses or CIDR networks, got %q", EnvTrustedProxies, item)
		}
		prefixes = append(prefixes, prefix)
	}
	*target = prefixes
	return nil
}

func parseNetwork(text string) (netip.Prefix, error) {
	if strings.Contains(text, "/") {
		prefix, err := netip.ParsePrefix(text)
		return prefix.Masked(), err
	}
	addr, err := netip.ParseAddr(text)
	if err != nil {
		return netip.Prefix{}, err
	}
	return netip.PrefixFrom(addr, addr.BitLen()), nil
}

func loadLogLevel(lookup LookupFunc, target *slog.Level) error {
	raw, ok := lookup(EnvLogLevel)
	if !ok {
		return nil
	}
	var level slog.Level
	if err := level.UnmarshalText([]byte(raw)); err != nil {
		return fmt.Errorf("config: %s must be one of debug, info, warn, error, got %q", EnvLogLevel, raw)
	}
	*target = level
	return nil
}
