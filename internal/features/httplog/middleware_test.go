package httplog

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestMiddleware_ExcludePaths(t *testing.T) {
	cfg := &Config{
		ExcludePaths:   []string{"/health", "/metrics"},
		RedactKeywords: []string{},
	}

	// Capture logs
	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
	slog.SetDefault(logger)

	handler := Middleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	tests := []struct {
		path      string
		shouldLog bool
	}{
		{"/health", false},
		{"/metrics", false},
		{"/api/users", true},
		{"/", true},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			logBuf.Reset()

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			hasLog := strings.Contains(logBuf.String(), "http request")

			if tt.shouldLog && !hasLog {
				t.Errorf("Expected log for path %s but got none", tt.path)
			}
			if !tt.shouldLog && hasLog {
				t.Errorf("Expected no log for path %s but got: %s", tt.path, logBuf.String())
			}
		})
	}
}

func TestMiddleware_LogLevels(t *testing.T) {
	cfg := &Config{
		ExcludePaths:   []string{},
		RedactKeywords: []string{},
	}

	tests := []struct {
		name          string
		statusCode    int
		expectedLevel string
	}{
		{"2xx success", 200, "INFO"},
		{"3xx redirect", 301, "INFO"},
		{"4xx client error", 404, "WARN"},
		{"4xx bad request", 400, "WARN"},
		{"5xx server error", 500, "ERROR"},
		{"5xx bad gateway", 502, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logBuf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
			slog.SetDefault(logger)

			handler := Middleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if !strings.Contains(logBuf.String(), `"level":"`+tt.expectedLevel+`"`) {
				t.Errorf("Expected log level %s for status %d, got: %s",
					tt.expectedLevel, tt.statusCode, logBuf.String())
			}
		})
	}
}

func TestMiddleware_LogType(t *testing.T) {
	cfg := &Config{
		ExcludePaths:   []string{},
		RedactKeywords: []string{},
	}

	tests := []struct {
		statusCode   int
		expectedType string
	}{
		{200, "http.access"},
		{404, "http.access"},
		{500, "http.error"},
		{502, "http.error"},
	}

	for _, tt := range tests {
		t.Run(tt.expectedType, func(t *testing.T) {
			var logBuf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
			slog.SetDefault(logger)

			handler := Middleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			if !strings.Contains(logBuf.String(), `"type":"`+tt.expectedType+`"`) {
				t.Errorf("Expected log type %s for status %d, got: %s",
					tt.expectedType, tt.statusCode, logBuf.String())
			}
		})
	}
}

func TestMiddleware_TraceIDExtraction(t *testing.T) {
	cfg := &Config{
		ExcludePaths:   []string{},
		RedactKeywords: []string{},
	}

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
	slog.SetDefault(logger)

	handler := Middleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	expectedTraceID := "0af7651916cd43dd8448eb211c80319c"
	traceparent := "00-" + expectedTraceID + "-b7ad6b7169203331-01"

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("traceparent", traceparent)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !strings.Contains(logBuf.String(), `"trace_id":"`+expectedTraceID+`"`) {
		t.Errorf("Expected trace_id %s in log, got: %s", expectedTraceID, logBuf.String())
	}
}

func TestMiddleware_ClientIDAnonymous(t *testing.T) {
	cfg := &Config{
		ExcludePaths:   []string{},
		RedactKeywords: []string{},
	}

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
	slog.SetDefault(logger)

	handler := Middleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if !strings.Contains(logBuf.String(), `"client_id":"anonymous"`) {
		t.Errorf("Expected client_id to be anonymous, got: %s", logBuf.String())
	}
}

func TestMiddleware_HTTPMetrics(t *testing.T) {
	cfg := &Config{
		ExcludePaths:   []string{},
		RedactKeywords: []string{},
	}

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
	slog.SetDefault(logger)

	handler := Middleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("response body"))
	}))

	req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("request body"))
	req.Header.Set("User-Agent", "test-agent")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	var logEntry map[string]interface{}
	if err := json.Unmarshal(logBuf.Bytes(), &logEntry); err != nil {
		t.Fatalf("Failed to parse log JSON: %v", err)
	}

	httpGroup, ok := logEntry["http"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected 'http' group in log")
	}

	// Check HTTP metrics
	if httpGroup["method"] != "POST" {
		t.Errorf("Expected method POST, got %v", httpGroup["method"])
	}
	if httpGroup["path"] != "/test" {
		t.Errorf("Expected path /test, got %v", httpGroup["path"])
	}
	if httpGroup["status"] != float64(200) {
		t.Errorf("Expected status 200, got %v", httpGroup["status"])
	}
	if httpGroup["user_agent"] != "test-agent" {
		t.Errorf("Expected user_agent test-agent, got %v", httpGroup["user_agent"])
	}

	// Check duration_ms exists and is > 0
	if duration, ok := httpGroup["duration_ms"].(float64); !ok || duration < 0 {
		t.Errorf("Expected valid duration_ms, got %v", httpGroup["duration_ms"])
	}
}

func TestMiddleware_DebugMode(t *testing.T) {
	// Set LOG_LEVEL to debug
	os.Setenv("LOG_LEVEL", "debug")
	defer os.Unsetenv("LOG_LEVEL")

	cfg := &Config{
		ExcludePaths:   []string{},
		RedactKeywords: []string{"password"},
		Debug: DebugConfig{
			LogRequestHeaders: true,
			LogRequestBody:    true,
			MaxBodySize:       4096,
		},
	}

	var logBuf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
	slog.SetDefault(logger)

	handler := Middleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read body to ensure it's restored
		body, _ := io.ReadAll(r.Body)
		if string(body) != `{"password":"secret"}` {
			t.Error("Request body was not properly restored")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodPost, "/test",
		strings.NewReader(`{"password":"secret"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token123")
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	logOutput := logBuf.String()

	// Check that headers are logged
	if !strings.Contains(logOutput, "request_headers") {
		t.Error("Expected request_headers in debug mode")
	}

	// Check that body is logged with redaction
	if !strings.Contains(logOutput, "request_body") {
		t.Error("Expected request_body in debug mode")
	}
	if !strings.Contains(logOutput, "[REDACTED]") {
		t.Error("Expected password to be redacted in body")
	}
}

func TestMiddleware_BodyCaptureOnlyForJSON(t *testing.T) {
	os.Setenv("LOG_LEVEL", "debug")
	defer os.Unsetenv("LOG_LEVEL")

	cfg := &Config{
		ExcludePaths:   []string{},
		RedactKeywords: []string{},
		Debug: DebugConfig{
			LogRequestBody: true,
			MaxBodySize:    4096,
		},
	}

	tests := []struct {
		name        string
		contentType string
		shouldLog   bool
	}{
		{"JSON content", "application/json", true},
		{"Text content", "text/plain", true},
		{"Binary content", "application/octet-stream", false},
		{"Image content", "image/png", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var logBuf bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logBuf, nil))
			slog.SetDefault(logger)

			handler := Middleware(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader("body content"))
			req.Header.Set("Content-Type", tt.contentType)
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			hasBody := strings.Contains(logBuf.String(), "request_body")

			if tt.shouldLog && !hasBody {
				t.Errorf("Expected body to be logged for %s", tt.contentType)
			}
			if !tt.shouldLog && hasBody {
				t.Errorf("Expected body NOT to be logged for %s", tt.contentType)
			}
		})
	}
}

func TestResponseWriter_StatusCapture(t *testing.T) {
	tests := []struct {
		name           string
		writeHeader    bool
		expectedStatus int
	}{
		{"explicit 404", true, 404},
		{"default 200", false, 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rw := &responseWriter{
				ResponseWriter: httptest.NewRecorder(),
				status:         200,
			}

			if tt.writeHeader {
				rw.WriteHeader(tt.expectedStatus)
			}

			if rw.status != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rw.status)
			}
		})
	}
}

func TestResponseWriter_SizeCapture(t *testing.T) {
	rw := &responseWriter{
		ResponseWriter: httptest.NewRecorder(),
		status:         200,
	}

	data := []byte("test response body")
	n, err := rw.Write(data)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}

	if rw.size != len(data) {
		t.Errorf("Expected size %d, got %d", len(data), rw.size)
	}
}
