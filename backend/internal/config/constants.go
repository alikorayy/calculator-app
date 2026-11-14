package config

import (
	"log/slog"
	"time"
)

// Environment variable names read by Load.
const (
	EnvPort               = "PORT"
	EnvReadTimeout        = "READ_TIMEOUT"
	EnvWriteTimeout       = "WRITE_TIMEOUT"
	EnvIdleTimeout        = "IDLE_TIMEOUT"
	EnvShutdownTimeout    = "SHUTDOWN_TIMEOUT"
	EnvRequestTimeout     = "REQUEST_TIMEOUT"
	EnvLogLevel           = "LOG_LEVEL"
	EnvMaxBodyBytes       = "MAX_BODY_BYTES"
	EnvRateLimitPerSecond = "RATE_LIMIT_RPS"
	EnvRateLimitBurst     = "RATE_LIMIT_BURST"
	EnvTrustedProxies     = "TRUSTED_PROXIES"
)

// Default values used when an environment variable is not set.
const (
	DefaultPort            = 8080
	DefaultReadTimeout     = 5 * time.Second
	DefaultWriteTimeout    = 10 * time.Second
	DefaultIdleTimeout     = 60 * time.Second
	DefaultShutdownTimeout = 10 * time.Second
	DefaultRequestTimeout  = 5 * time.Second
	DefaultLogLevel        = slog.LevelInfo
	// DefaultMaxBodyBytes (4 KiB) is far above the largest valid request: an
	// expression is at most 256 characters.
	DefaultMaxBodyBytes = 4 << 10
	// A person using the calculator sends one request per "=", so these
	// limits only ever slow down scripts.
	DefaultRateLimitPerSecond = 10
	DefaultRateLimitBurst     = 20
)

// trustedProxySeparator separates the entries of TRUSTED_PROXIES.
const trustedProxySeparator = ","

// Valid TCP port range.
const (
	minPort = 1
	maxPort = 65535
)
