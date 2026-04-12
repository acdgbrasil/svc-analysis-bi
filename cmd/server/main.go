package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/acdgbrasil/svc-analysis-bi/configs"
	"github.com/acdgbrasil/svc-analysis-bi/internal/api"
	"github.com/acdgbrasil/svc-analysis-bi/internal/export"
	"github.com/acdgbrasil/svc-analysis-bi/internal/store"
)

func main() {
	os.Exit(run())
}

func run() int {
	// 1. Load configuration
	cfg, err := configs.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config: %v\n", err)
		return 1
	}

	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "config validation: %v\n", err)
		return 1
	}

	// 2. Setup structured logger
	var logLevel slog.Level
	switch cfg.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: logLevel,
	}))

	logger.Info("starting svc-analysis-bi",
		"port", cfg.Server.Port,
		"host", cfg.Server.Host,
		"log_level", cfg.LogLevel,
	)

	// 3. Connect to database
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := store.New(ctx, cfg.Database.DSN(), cfg.Database.MaxConns)
	if err != nil {
		logger.Error("failed to connect to database", "error", err)
		return 1
	}
	defer db.Close()

	logger.Info("database connected")

	// 4. Create stores
	indicatorStore := store.NewIndicatorStore(db.Pool())

	// 5. Create export encoder registry
	encoders := export.NewRegistry()

	// 6. Build router
	router := api.NewRouter(api.RouterDeps{
		Logger:         logger,
		DB:             db,
		Indicators:     indicatorStore,
		Encoders:       encoders,
		RateLimitRPS:   100,
		RateLimitBurst: 200,
		// JWTValidator and NATS will be wired when those adapters are built.
		// For now they are nil, meaning JWT auth is disabled and NATS readiness
		// always reports as connected.
	})

	// 7. Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	srv := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// 8. Graceful shutdown setup
	shutdownCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// 9. Start server
	errCh := make(chan error, 1)
	go func() {
		logger.Info("HTTP server listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	// 10. Wait for shutdown signal or server error
	select {
	case err := <-errCh:
		if err != nil {
			logger.Error("server error", "error", err)
			return 1
		}
	case <-shutdownCtx.Done():
		logger.Info("shutdown signal received, draining connections...")
		drainCtx, drainCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer drainCancel()

		if err := srv.Shutdown(drainCtx); err != nil {
			logger.Error("server shutdown error", "error", err)
			return 1
		}
		logger.Info("server shut down gracefully")
	}

	return 0
}
