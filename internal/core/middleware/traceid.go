package middleware

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"net/http"
)

const traceParentHeader = "traceparent"

// TraceID middleware ensures every request has a W3C-compliant traceparent header.
// If the header is already present, it propagates it; otherwise, it generates a new one.
// This enables request correlation across services and log aggregation.
func TraceID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceparent := r.Header.Get(traceParentHeader)

			if traceparent == "" {
				traceparent = generateTraceParent()
				r.Header.Set(traceParentHeader, traceparent)

				slog.Debug("traceparent generated",
					"traceparent", traceparent,
					"path", r.URL.Path)
			} else {
				slog.Debug("traceparent propagated",
					"traceparent", traceparent,
					"path", r.URL.Path)
			}

			next.ServeHTTP(w, r)
		})
	}
}

// generateTraceParent creates a W3C-compliant traceparent header.
// Format: version-trace_id-parent_id-trace_flags
// Example: 00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01
func generateTraceParent() string {
	traceID := make([]byte, 16)
	if _, err := rand.Read(traceID); err != nil {
		slog.Warn("failed to generate trace ID, using zeros", "error", err)
	}

	spanID := make([]byte, 8)
	if _, err := rand.Read(spanID); err != nil {
		slog.Warn("failed to generate span ID, using zeros", "error", err)
	}

	return fmt.Sprintf("00-%x-%x-01", traceID, spanID)
}
