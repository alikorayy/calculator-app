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
	LogLevel        slog.Level
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
		LogLevel:        DefaultLogLevel,
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
	}
	for _, d := range durations {
		if err := loadDuration(lookup, d.key, d.target); err != nil {
			return Config{}, err
		}
	}

	if err := loadLogLevel(lookup, &cfg.LogLevel); err != nil {
		return Config{}, err
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
