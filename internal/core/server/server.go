package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/noyellowline/gokart/internal/core/config"
	"github.com/noyellowline/gokart/internal/core/proxy"
)

type Server struct {
	httpServer *http.Server
	cfg        *config.ServerConfig
	proxy      *proxy.Proxy
}

func New(cfg *config.Config) (*Server, error) {
	// Initialize proxy
	proxyHandler, err := proxy.New(&cfg.Core.Proxy)
	if err != nil {
		return nil, err
	}

	// Setup routes (health + proxied routes with middleware + features)
	handler := setupRoutes(proxyHandler, cfg)

	httpServer := &http.Server{
		Addr:           cfg.Core.Server.Addr,
		Handler:        handler,
		ReadTimeout:    cfg.Core.Server.ReadTimeout,
		WriteTimeout:   cfg.Core.Server.WriteTimeout,
		IdleTimeout:    cfg.Core.Server.IdleTimeout,
		MaxHeaderBytes: cfg.Core.Server.MaxHeaderBytes,
	}

	return &Server{
		httpServer: httpServer,
		cfg:        &cfg.Core.Server,
		proxy:      proxyHandler,
	}, nil
}

func (s *Server) Start() error {
	slog.Info("http server starting", "addr", s.cfg.Addr)

	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("http server shutting down")

	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		slog.Error("http server shutdown failed", "error", err)
		return err
	}

	slog.Info("http server stopped")
	return nil
}
