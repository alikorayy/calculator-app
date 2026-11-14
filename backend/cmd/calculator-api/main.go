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

	"github.com/alikorayy/calculator-app/backend/internal/config"
	"github.com/alikorayy/calculator-app/backend/internal/httpapi"
)

func main() {
	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		// The logger is configured from cfg, so a config error is reported on stderr directly.
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))

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

	server := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      httpapi.NewRouter(),
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
