package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/noyellowline/gokart/internal/core/config"
	"github.com/noyellowline/gokart/internal/core/logging"
	"github.com/noyellowline/gokart/internal/core/server"
)

const version = "0.1.0"

func main() {
	// Setup sidecar internal logging
	logging.Setup()

	slog.Info("starting gokart sidecar", "version", version)

	// Load configuration
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = config.DefaultConfigPath
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		slog.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	slog.Info("configuration loaded successfully",
		"server_addr", cfg.Core.Server.Addr,
		"proxy_target", cfg.Core.Proxy.Target)

	// Create HTTP server with proxy
	srv, err := server.New(cfg)
	if err != nil {
		slog.Error("failed to initialize server", "error", err)
		os.Exit(1)
	}

	// Start server in background
	errCh := make(chan error, 1)
	go func() {
		if err := srv.Start(); err != nil {
			errCh <- err
		}
	}()

	slog.Info("sidecar started successfully")

	// Wait for interrupt signal or server error
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		slog.Error("server error", "error", err)
		os.Exit(1)
	case sig := <-sigCh:
		slog.Info("shutdown signal received", "signal", sig)
	}

	// Graceful shutdown
	slog.Info("shutting down gracefully...")
	if err := srv.Shutdown(context.Background()); err != nil {
		slog.Error("shutdown error", "error", err)
		os.Exit(1)
	}

	slog.Info("sidecar stopped")
}
