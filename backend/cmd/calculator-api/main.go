// Command calculator-api runs the calculator HTTP service.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alikorayy/calculator-app/backend/internal/config"
	"github.com/alikorayy/calculator-app/backend/internal/httpapi"
	"github.com/alikorayy/calculator-app/backend/internal/middleware"
)

// How often idle clients are removed from the rate limiter, and after how
// long without a request a client counts as idle.
const (
	rateLimitCleanupInterval = time.Minute
	rateLimitIdleAfter       = 3 * time.Minute
)

func main() {
	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		// The logger is configured from cfg, so a config error is reported on stderr directly.
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if len(os.Args) > 1 && os.Args[1] == healthcheckCommand {
		if err := checkHealth(healthcheckURL(cfg.Port)); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	jsonHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel})
	logger := slog.New(middleware.NewRequestIDLogHandler(jsonHandler))

	if err := run(cfg, logger); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

// run starts the HTTP server and blocks until it fails or a SIGINT/SIGTERM arrives.
// On a signal it stops accepting new connections and waits up to cfg.ShutdownTimeout
// for in-flight requests to finish.
func run(cfg config.Config, logger *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	limiter := newRateLimiter(cfg)
	go limiter.Cleanup(ctx, rateLimitCleanupInterval, rateLimitIdleAfter)

	server := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      newHandler(cfg, logger, limiter),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
		ErrorLog:     slog.NewLogLogger(logger.Handler(), slog.LevelError),
	}

	serverErr := make(chan error, 1)
	go func() {
		logger.Info("server starting", "addr", server.Addr)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("listen: %w", err)
	case <-ctx.Done():
	}

	// Restore default signal handling so a second Ctrl+C terminates immediately.
	stop()
	logger.Info("shutdown started", "timeout", cfg.ShutdownTimeout.String())

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}

	logger.Info("shutdown complete")
	return nil
}

func newRateLimiter(cfg config.Config) *middleware.RateLimiter {
	return middleware.NewRateLimiter(float64(cfg.RateLimitPerSecond), int(cfg.RateLimitBurst), cfg.TrustedProxies)
}

// newHandler wires the API router into the middleware chain. The order matters:
//
//   - RequestID first, so every later layer and every log line has the ID;
//   - Logging next, so it records the final status of everything below it,
//     including a 500 produced by Recover;
//   - Recover, so a panic anywhere below becomes a 500 instead of a dropped
//     connection;
//   - SecurityHeaders before any layer that can answer early, so 429, 413 and
//     503 responses carry them too;
//   - RateLimit before reading the body, so rejected clients cost almost nothing;
//   - MaxBodySize and Timeout closest to the handlers they protect.
func newHandler(cfg config.Config, logger *slog.Logger, limiter *middleware.RateLimiter) http.Handler {
	return middleware.Chain(httpapi.NewRouter(logger),
		middleware.RequestID,
		middleware.Logging(logger),
		middleware.Recover(logger, httpapi.WriteError),
		middleware.SecurityHeaders,
		limiter.Middleware(httpapi.WriteError),
		middleware.MaxBodySize(cfg.MaxBodyBytes, httpapi.WriteError),
		middleware.Timeout(cfg.RequestTimeout, httpapi.WriteError),
	)
}
