package httplog

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"
)

// Context keys for feature enrichment
type contextKey string

const (
	ContextKeyClientID contextKey = "client_id"
)

// Log type constants
const (
	LogTypeHTTPAccess = "http.access"
	LogTypeHTTPError  = "http.error"
)

// responseWriter wraps http.ResponseWriter to capture status and size.
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
	body   *bytes.Buffer // Capture body if debug mode
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size

	// Capture body if debug enabled
	if rw.body != nil && size > 0 {
		rw.body.Write(b[:size])
	}

	return size, err
}

// Middleware creates the HTTP logging middleware.
func Middleware(cfg *Config) func(http.Handler) http.Handler {
	redactor := NewRedactor(cfg.RedactKeywords)
	isDebug := isDebugLevel()

	// Get service_id from environment
	serviceID := os.Getenv("SERVICE_ID")
	if serviceID == "" {
		serviceID = "unknown-service"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip excluded paths (fast return, zero allocation)
			if shouldSkip(r.URL.Path, cfg.ExcludePaths) {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()

			// Prepare response writer with optional body capture
			var bodyBuf *bytes.Buffer
			if isDebug && cfg.Debug.LogResponseBody {
				bodyBuf = &bytes.Buffer{}
			}

			wrapped := &responseWriter{
				ResponseWriter: w,
				status:         200,
				body:           bodyBuf,
			}

			// Capture request body if debug mode
			var reqBody []byte
			if isDebug && cfg.Debug.LogRequestBody {
				reqBody = captureRequestBody(r, cfg.Debug.MaxBodySize)
			}

			// Execute request
			next.ServeHTTP(wrapped, r)

			duration := time.Since(start)

			// Extract trace_id from traceparent header (W3C format)
			traceID := extractTraceID(r.Header.Get("traceparent"))

			// Extract client_id from context (added by auth feature, future)
			clientID := extractClientID(r)

			// Build log attributes
			attrs := []slog.Attr{
				slog.String("type", logType(wrapped.status)),
				slog.String("trace_id", traceID),
				slog.String("service_id", serviceID),
				slog.String("client_id", clientID),
				slog.Group("http",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.Int("status", wrapped.status),
					slog.Int64("duration_ms", duration.Milliseconds()),
					slog.Int("request_size", int(r.ContentLength)),
					slog.Int("response_size", wrapped.size),
					slog.String("remote_addr", r.RemoteAddr),
					slog.String("user_agent", r.UserAgent()),
				),
			}

			// Add debug information (headers, body) if LOG_LEVEL=debug
			if isDebug {
				if cfg.Debug.LogRequestHeaders {
					attrs = append(attrs, slog.Any("request_headers",
						redactor.RedactHeaders(r.Header)))
				}

				if cfg.Debug.LogRequestBody && len(reqBody) > 0 {
					attrs = append(attrs, slog.String("request_body",
						redactor.RedactBody(string(reqBody))))
				}

				if cfg.Debug.LogResponseBody && bodyBuf != nil && bodyBuf.Len() > 0 {
					attrs = append(attrs, slog.String("response_body",
						redactor.RedactBody(bodyBuf.String())))
				}
			}

			// Log with appropriate level based on status code
			level := logLevelForStatus(wrapped.status)
			slog.LogAttrs(r.Context(), level, "http request", attrs...)
		})
	}
}

// logType returns the log type based on response status.
func logType(status int) string {
	if status >= 500 {
		return LogTypeHTTPError
	}
	return LogTypeHTTPAccess
}

// logLevelForStatus returns the appropriate log level for a status code.
func logLevelForStatus(status int) slog.Level {
	switch {
	case status >= 500:
		return slog.LevelError
	case status >= 400:
		return slog.LevelWarn
	default:
		return slog.LevelInfo
	}
}

// extractTraceID extracts the trace_id from W3C traceparent header.
// Format: 00-{trace_id}-{span_id}-01
func extractTraceID(traceparent string) string {
	if len(traceparent) < 36 {
		return ""
	}

	parts := strings.Split(traceparent, "-")
	if len(parts) >= 2 {
		return parts[1] // Return 32-char trace_id
	}

	return ""
}

// extractClientID extracts client_id from request context.
// Returns "anonymous" if not present (before auth feature is implemented).
func extractClientID(r *http.Request) string {
	if clientID := r.Context().Value(ContextKeyClientID); clientID != nil {
		if id, ok := clientID.(string); ok {
			return id
		}
	}
	return "anonymous"
}

// captureRequestBody reads and captures the request body up to maxSize.
// Restores the body for downstream handlers.
func captureRequestBody(r *http.Request, maxSize int) []byte {
	if r.Body == nil || r.ContentLength == 0 {
		return nil
	}

	// Only capture JSON and text content types
	contentType := r.Header.Get("Content-Type")
	if !isCapturableContentType(contentType) {
		return nil
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, int64(maxSize)))
	if err != nil {
		return nil
	}

	// Restore body for downstream handlers
	r.Body = io.NopCloser(io.MultiReader(
		bytes.NewReader(body),
		r.Body,
	))

	return body
}

// isCapturableContentType checks if content type should be captured.
func isCapturableContentType(contentType string) bool {
	contentType = strings.ToLower(contentType)
	return strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "text/")
}

// shouldSkip checks if a path should be excluded from logging.
func shouldSkip(path string, excludePaths []string) bool {
	for _, excluded := range excludePaths {
		if path == excluded {
			return true
		}
	}
	return false
}

// isDebugLevel checks if current log level is debug.
func isDebugLevel() bool {
	logLevel := strings.ToLower(os.Getenv("LOG_LEVEL"))
	return logLevel == "debug"
}
