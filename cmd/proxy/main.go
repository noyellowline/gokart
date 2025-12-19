package main

import (
	"log/slog"
	"os"

	"github.com/noyellowline/gokart/internal/core/config"
	"github.com/noyellowline/gokart/internal/core/logging"
)

const version = "0.1.0"

func main() {
	// Setup sidecarinternal logging
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

	// TODO: Initialize features
	// TODO: Setup proxy
	// TODO: Start server

	slog.Info("sidecar started successfully")
}
