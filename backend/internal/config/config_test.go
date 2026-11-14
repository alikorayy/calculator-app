package config

import (
	"log/slog"
	"net/netip"
	"reflect"
	"testing"
	"time"
)

// lookupFrom returns a LookupFunc backed by a map, so tests never touch the real environment.
func lookupFrom(env map[string]string) LookupFunc {
	return func(key string) (string, bool) {
		v, ok := env[key]
		return v, ok
	}
}

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load(lookupFrom(nil))
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	want := Config{
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
	if !reflect.DeepEqual(cfg, want) {
		t.Errorf("Load() = %+v, want %+v", cfg, want)
	}
}

func TestLoad_Overrides(t *testing.T) {
	env := map[string]string{
		EnvPort:               "9090",
		EnvReadTimeout:        "1s",
		EnvWriteTimeout:       "2s",
		EnvIdleTimeout:        "3m",
		EnvShutdownTimeout:    "500ms",
		EnvRequestTimeout:     "1500ms",
		EnvLogLevel:           "debug",
		EnvMaxBodyBytes:       "1024",
		EnvRateLimitPerSecond: "5",
		EnvRateLimitBurst:     "8",
		EnvTrustedProxies:     " 172.16.0.0/12, 10.0.0.5 ,, ::1 ",
	}

	cfg, err := Load(lookupFrom(env))
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}

	want := Config{
		Port:               9090,
		ReadTimeout:        time.Second,
		WriteTimeout:       2 * time.Second,
		IdleTimeout:        3 * time.Minute,
		ShutdownTimeout:    500 * time.Millisecond,
		RequestTimeout:     1500 * time.Millisecond,
		LogLevel:           slog.LevelDebug,
		MaxBodyBytes:       1024,
		RateLimitPerSecond: 5,
		RateLimitBurst:     8,
		TrustedProxies: []netip.Prefix{
			netip.MustParsePrefix("172.16.0.0/12"),
			netip.MustParsePrefix("10.0.0.5/32"),
			netip.MustParsePrefix("::1/128"),
		},
	}
	if !reflect.DeepEqual(cfg, want) {
		t.Errorf("Load() = %+v, want %+v", cfg, want)
	}
}

func TestLoad_InvalidValues(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"port not a number", EnvPort, "abc"},
		{"port zero", EnvPort, "0"},
		{"port too large", EnvPort, "70000"},
		{"port negative", EnvPort, "-1"},
		{"read timeout not a duration", EnvReadTimeout, "fast"},
		{"write timeout without unit", EnvWriteTimeout, "10"},
		{"idle timeout zero", EnvIdleTimeout, "0s"},
		{"shutdown timeout negative", EnvShutdownTimeout, "-5s"},
		{"request timeout not a duration", EnvRequestTimeout, "soon"},
		{"unknown log level", EnvLogLevel, "verbose"},
		{"body limit not a number", EnvMaxBodyBytes, "4KB"},
		{"body limit zero", EnvMaxBodyBytes, "0"},
		{"body limit negative", EnvMaxBodyBytes, "-1"},
		{"rate not a number", EnvRateLimitPerSecond, "fast"},
		{"burst zero", EnvRateLimitBurst, "0"},
		{"trusted proxy not an address", EnvTrustedProxies, "nginx"},
		{"trusted proxy with a bad prefix length", EnvTrustedProxies, "10.0.0.0/33"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Load(lookupFrom(map[string]string{tt.key: tt.value}))
			if err == nil {
				t.Errorf("Load() with %s=%q: expected an error, got nil", tt.key, tt.value)
			}
		})
	}
}

func TestConfig_Addr(t *testing.T) {
	cfg := Config{Port: 8080}
	if got, want := cfg.Addr(), ":8080"; got != want {
		t.Errorf("Addr() = %q, want %q", got, want)
	}
}

func TestLoad_RequestTimeoutMustBeShorterThanWriteTimeout(t *testing.T) {
	tests := []struct {
		name           string
		requestTimeout string
		writeTimeout   string
	}{
		{"equal", "10s", "10s"},
		{"longer", "15s", "10s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := map[string]string{EnvRequestTimeout: tt.requestTimeout, EnvWriteTimeout: tt.writeTimeout}
			if _, err := Load(lookupFrom(env)); err == nil {
				t.Errorf("Load() with %s=%s and %s=%s: expected an error, got nil",
					EnvRequestTimeout, tt.requestTimeout, EnvWriteTimeout, tt.writeTimeout)
			}
		})
	}
}

func TestLoad_TrustedProxyNetworksAreNormalized(t *testing.T) {
	cfg, err := Load(lookupFrom(map[string]string{EnvTrustedProxies: "172.18.0.7/16"}))
	if err != nil {
		t.Fatalf("Load() unexpected error: %v", err)
	}
	want := []netip.Prefix{netip.MustParsePrefix("172.18.0.0/16")}
	if !reflect.DeepEqual(cfg.TrustedProxies, want) {
		t.Errorf("TrustedProxies = %v, want %v", cfg.TrustedProxies, want)
	}
}
