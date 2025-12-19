package httplog

import (
	"net/http"
	"testing"
)

func TestRedactHeaders(t *testing.T) {
	tests := []struct {
		name     string
		keywords []string
		headers  http.Header
		want     map[string]string
	}{
		{
			name:     "redact authorization header",
			keywords: []string{"authorization"},
			headers: http.Header{
				"Authorization": []string{"Bearer token123"},
				"Content-Type":  []string{"application/json"},
			},
			want: map[string]string{
				"Authorization": "[REDACTED]",
				"Content-Type":  "application/json",
			},
		},
		{
			name:     "redact multiple sensitive headers",
			keywords: []string{"authorization", "cookie", "api_key"},
			headers: http.Header{
				"Authorization": []string{"Bearer token123"},
				"Cookie":        []string{"session=abc123"},
				"X-Api-Key":     []string{"sk_live_123"},
				"User-Agent":    []string{"curl/7.68.0"},
			},
			want: map[string]string{
				"Authorization": "[REDACTED]",
				"Cookie":        "[REDACTED]",
				"X-Api-Key":     "[REDACTED]",
				"User-Agent":    "curl/7.68.0",
			},
		},
		{
			name:     "case insensitive matching",
			keywords: []string{"authorization"},
			headers: http.Header{
				"AUTHORIZATION": []string{"Bearer token123"},
				"authorization": []string{"Bearer token456"},
			},
			want: map[string]string{
				"AUTHORIZATION": "[REDACTED]",
				"authorization": "[REDACTED]",
			},
		},
		{
			name:     "partial keyword matching in header name",
			keywords: []string{"token"},
			headers: http.Header{
				"X-Auth-Token":    []string{"secret123"},
				"X-Refresh-Token": []string{"refresh456"},
				"Content-Type":    []string{"application/json"},
			},
			want: map[string]string{
				"X-Auth-Token":    "[REDACTED]",
				"X-Refresh-Token": "[REDACTED]",
				"Content-Type":    "application/json",
			},
		},
		{
			name:     "no redaction when disabled",
			keywords: []string{},
			headers: http.Header{
				"Authorization": []string{"Bearer token123"},
				"Cookie":        []string{"session=abc123"},
			},
			want: map[string]string{
				"Authorization": "Bearer token123",
				"Cookie":        "session=abc123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redactor := NewRedactor(tt.keywords)
			got := redactor.RedactHeaders(tt.headers)

			for key, want := range tt.want {
				if got[key] != want {
					t.Errorf("RedactHeaders()[%s] = %v, want %v", key, got[key], want)
				}
			}
		})
	}
}

func TestRedactBody(t *testing.T) {
	tests := []struct {
		name     string
		keywords []string
		body     string
		want     string
	}{
		{
			name:     "redact password in JSON",
			keywords: []string{"password"},
			body:     `{"username":"john","password":"secret123"}`,
			want:     `{"password":"[REDACTED]","username":"john"}`,
		},
		{
			name:     "redact multiple sensitive fields",
			keywords: []string{"password", "token", "api_key"},
			body:     `{"username":"john","password":"secret123","api_key":"sk_live_123","token":"abc"}`,
			want:     `{"api_key":"[REDACTED]","password":"[REDACTED]","token":"[REDACTED]","username":"john"}`,
		},
		{
			name:     "redact nested objects",
			keywords: []string{"password", "secret"},
			body:     `{"user":{"email":"john@test.com","password":"secret123"},"config":{"secret":"api_secret"}}`,
			want:     `{"config":{"secret":"[REDACTED]"},"user":{"email":"john@test.com","password":"[REDACTED]"}}`,
		},
		{
			name:     "case insensitive field matching",
			keywords: []string{"password"},
			body:     `{"Password":"secret123","PASSWORD":"secret456","username":"john"}`,
			want:     `{"PASSWORD":"[REDACTED]","Password":"[REDACTED]","username":"john"}`,
		},
		{
			name:     "partial keyword matching in field name",
			keywords: []string{"token"},
			body:     `{"auth_token":"abc123","refresh_token":"def456","email":"test@example.com"}`,
			want:     `{"auth_token":"[REDACTED]","email":"test@example.com","refresh_token":"[REDACTED]"}`,
		},
		{
			name:     "non-JSON body returns unchanged",
			keywords: []string{"password"},
			body:     "username=john&password=secret123",
			want:     "username=john&password=secret123",
		},
		{
			name:     "empty body returns empty",
			keywords: []string{"password"},
			body:     "",
			want:     "",
		},
		{
			name:     "no redaction when disabled",
			keywords: []string{},
			body:     `{"password":"secret123","token":"abc"}`,
			want:     `{"password":"secret123","token":"abc"}`,
		},
		{
			name:     "redact in arrays of objects",
			keywords: []string{"password"},
			body:     `{"users":[{"name":"john","password":"secret1"},{"name":"jane","password":"secret2"}]}`,
			want:     `{"users":[{"name":"john","password":"[REDACTED]"},{"name":"jane","password":"[REDACTED]"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redactor := NewRedactor(tt.keywords)
			got := redactor.RedactBody(tt.body)

			if got != tt.want {
				t.Errorf("RedactBody() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRedactorDisabled(t *testing.T) {
	// Test that redactor with no keywords is disabled
	redactor := NewRedactor(nil)

	if redactor.enabled {
		t.Error("Redactor should be disabled when keywords are nil")
	}

	redactor = NewRedactor([]string{})
	if redactor.enabled {
		t.Error("Redactor should be disabled when keywords are empty")
	}
}
