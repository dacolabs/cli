package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dacolabs/daco/internal/api"
	"github.com/dacolabs/daco/internal/telemetry"
	"github.com/dacolabs/daco/internal/version"
	"github.com/lmittmann/tint"
)

// Run is the main application logic, extracted for testability.
// It accepts OS dependencies as parameters (context, env lookup).
func Run(ctx context.Context, getenv func(string) string) error {
	// Load configuration
	cfg, err := LoadConfigWithEnv(getenv)
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	// Setup telemetry
	otelShutdown, err := telemetry.SetupOpenTelemetry(ctx)
	if err != nil {
		return fmt.Errorf("setup opentelemetry: %w", err)
	}
	defer func() {
		if err := otelShutdown(ctx); err != nil {
			slog.Error("opentelemetry shutdown failed", "error", err)
		}
	}()

	// Setup logging based on format
	if cfg.Log.Format == LogFormatText {
		slog.SetDefault(slog.New(tint.NewHandler(os.Stderr, &tint.Options{
			Level: slog.LevelDebug,
		})))
	} else {
		telemetry.SetupSlog("daco-api");
	}

	slog.Info("starting daco api",
		"version", version.Version,
		"commit", version.Commit,
		"port", cfg.Server.Port,
	)

	// Create HTTP server with all dependencies
	handler := api.NewServer()

	// Create HTTP server with configured timeouts
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in goroutine
	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("http server listening", "port", cfg.Server.Port)
		serverErrors <- httpServer.ListenAndServe()
	}()

	// Wait for interrupt signal or server error
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		slog.Info("received shutdown signal", "signal", sig.String())

		// Graceful shutdown with configured timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			slog.Error("graceful shutdown failed", "error", err)
			if err := httpServer.Close(); err != nil {
				slog.Error("server close failed", "error", err)
			}
			return fmt.Errorf("shutdown: %w", err)
		}

		slog.Info("server stopped gracefully")
	}

	return nil
}
