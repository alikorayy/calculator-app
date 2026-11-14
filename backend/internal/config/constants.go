package config

import (
	"log/slog"
	"time"
)

// Environment variable names read by Load.
const (
	EnvPort            = "PORT"
	EnvReadTimeout     = "READ_TIMEOUT"
	EnvWriteTimeout    = "WRITE_TIMEOUT"
	EnvIdleTimeout     = "IDLE_TIMEOUT"
	EnvShutdownTimeout = "SHUTDOWN_TIMEOUT"
	EnvLogLevel        = "LOG_LEVEL"
)

// Default values used when an environment variable is not set.
const (
	DefaultPort            = 8080
	DefaultReadTimeout     = 5 * time.Second
	DefaultWriteTimeout    = 10 * time.Second
	DefaultIdleTimeout     = 60 * time.Second
	DefaultShutdownTimeout = 10 * time.Second
	DefaultLogLevel        = slog.LevelInfo
)

// Valid TCP port range.
const (
	minPort = 1
	maxPort = 65535
)
