package server

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/noyellowline/gokart/internal/api/health"
	"github.com/noyellowline/gokart/internal/core/middleware"
)

// setup HTTP routes: internal API endpoints + proxied routes
func setupRoutes(proxyHandler http.Handler) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", health.Handler())

	mux.Handle("/", setupMiddlewareChain(proxyHandler))

	return mux
}

// setupMiddlewareChain wraps the proxy handler with core middleware.
// Middleware are applied in reverse order (last defined = first executed).
func setupMiddlewareChain(proxyHandler http.Handler) http.Handler {
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

	slog.Info("middleware chain configured",
		"service_id", serviceID,
		"middleware", []string{"traceid", "internal_headers"})

	return handler
}
