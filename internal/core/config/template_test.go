package config

import (
	"os"
	"testing"
)

func TestExpandTemplates(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		envVars        map[string]string
		want           string
		wantUnexpanded []string
	}{
		{
			name:    "simple substitution",
			input:   "addr: ${HOST}:${PORT}",
			envVars: map[string]string{"HOST": "localhost", "PORT": "8080"},
			want:    "addr: localhost:8080",
		},
		{
			name:    "with default value",
			input:   "addr: ${HOST:0.0.0.0}:${PORT:8080}",
			envVars: map[string]string{},
			want:    "addr: 0.0.0.0:8080",
		},
		{
			name:    "env overrides default",
			input:   "level: ${LOG_LEVEL:info}",
			envVars: map[string]string{"LOG_LEVEL": "debug"},
			want:    "level: debug",
		},
		{
			name:           "missing required var",
			input:          "addr: ${REQUIRED_VAR}",
			envVars:        map[string]string{},
			want:           "addr: ${REQUIRED_VAR}",
			wantUnexpanded: []string{"REQUIRED_VAR"},
		},
		{
			name:    "mixed vars",
			input:   "url: ${PROTO:http}://${HOST}:${PORT:80}",
			envVars: map[string]string{"HOST": "example.com"},
			want:    "url: http://example.com:80",
		},
		{
			name:    "no templates",
			input:   "static: value",
			envVars: map[string]string{},
			want:    "static: value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set env vars
			for k, v := range tt.envVars {
				os.Setenv(k, v)
				defer os.Unsetenv(k)
			}

			got, unexpanded, err := expandTemplates([]byte(tt.input))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if string(got) != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}

			if len(unexpanded) != len(tt.wantUnexpanded) {
				t.Errorf("unexpanded count: got %d, want %d", len(unexpanded), len(tt.wantUnexpanded))
			}
			for _, want := range tt.wantUnexpanded {
				found := false
				for _, u := range unexpanded {
					if u == want {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("missing unexpanded var: %s", want)
				}
			}
		})
	}
}

func TestExpandTemplates_MaxDepth(t *testing.T) {
	// Create circular reference that would exceed maxDepth
	os.Setenv("VAR1", "${VAR2}")
	os.Setenv("VAR2", "${VAR1}")
	defer os.Unsetenv("VAR1")
	defer os.Unsetenv("VAR2")

	_, _, err := expandTemplates([]byte("test: ${VAR1}"))
	if err == nil {
		t.Error("expected error for exceeding max depth, got nil")
	}
}
