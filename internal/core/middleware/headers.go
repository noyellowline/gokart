package middleware

import (
	"log/slog"
	"net/http"
)

const (
	serviceIDHeader = "X-Service-ID"
	clientIDHeader  = "X-Client-ID"
)

// InternalHeaders middleware adds internal headers to requests forwarded to the backend app.
// These headers are used by the backend for logging, tracing, and identification purposes.
func InternalHeaders(serviceID string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if serviceID != "" {
				r.Header.Set(serviceIDHeader, serviceID)

				slog.Debug("internal headers set",
					"service_id", serviceID,
					"path", r.URL.Path)
			}

			next.ServeHTTP(w, r)
		})
	}
}
