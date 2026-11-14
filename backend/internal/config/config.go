// Package config loads the service configuration from environment variables.
package config

import (
	"fmt"
	"log/slog"
	"strconv"
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
}

// LookupFunc returns the value of an environment variable and whether it is set.
// os.LookupEnv satisfies it; tests pass a map-backed function instead.
type LookupFunc func(key string) (string, bool)

// Load builds a Config from the given lookup function, applying defaults for
// unset variables. It fails fast on any invalid value so that a misconfigured
// service never starts.
func Load(lookup LookupFunc) (Config, error) {
	cfg := Config{
		Port:            DefaultPort,
		ReadTimeout:     DefaultReadTimeout,
		WriteTimeout:    DefaultWriteTimeout,
		IdleTimeout:     DefaultIdleTimeout,
		ShutdownTimeout: DefaultShutdownTimeout,
		RequestTimeout:  DefaultRequestTimeout,
		LogLevel:        DefaultLogLevel,
		MaxBodyBytes:    DefaultMaxBodyBytes,
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
	if err := loadPositiveInt(lookup, EnvMaxBodyBytes, &cfg.MaxBodyBytes); err != nil {
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
