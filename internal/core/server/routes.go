package server

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/noyellowline/gokart/internal/api/health"
	"github.com/noyellowline/gokart/internal/core/config"
	"github.com/noyellowline/gokart/internal/core/middleware"
	"github.com/noyellowline/gokart/internal/features/httplog"
)

// setup HTTP routes: internal API endpoints + proxied routes
func setupRoutes(proxyHandler http.Handler, cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", health.Handler())

	mux.Handle("/", setupMiddlewareChain(proxyHandler, cfg))

	return mux
}

// setupMiddlewareChain wraps the proxy handler with core middleware and features.
// Middleware are applied in reverse order (last defined = first executed).
func setupMiddlewareChain(proxyHandler http.Handler, cfg *config.Config) http.Handler {
	handler := proxyHandler

	serviceID := os.Getenv("SERVICE_ID")
	if serviceID == "" {
		serviceID = "unknown-service"
		slog.Warn("SERVICE_ID not set, using default", "service_id", serviceID)
	}

	// Add internal headers
	handler = middleware.InternalHeaders(serviceID)(handler)

	// Trace ID (ensures traceparent header, W3C compliant)
	handler = middleware.TraceID()(handler)

	// HTTP Logging feature (if enabled)
	if featureCfg, exists := cfg.Features["http_logging"]; exists && featureCfg.Enabled {
		httpLogFeature, err := httplog.New(featureCfg)
		if err != nil {
			slog.Error("failed to initialize http_logging feature", "error", err)
		} else {
			handler = httpLogFeature.Middleware()(handler)
			slog.Info("http_logging feature enabled")
		}
	}

	slog.Info("middleware chain configured",
		"service_id", serviceID,
		"middleware", []string{"traceid", "internal_headers", "http_logging"})

	return handler
}
