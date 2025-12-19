package httplog

import (
	"testing"

	"github.com/noyellowline/gokart/internal/core/config"
	"gopkg.in/yaml.v3"
)

func TestConfig_ApplyDefaults(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		want   Config
	}{
		{
			name: "apply default max body size",
			config: Config{
				Debug: DebugConfig{
					MaxBodySize: 0, // Not set
				},
			},
			want: Config{
				Debug: DebugConfig{
					MaxBodySize: 4096,
				},
				RedactKeywords: []string{
					"password", "token", "secret",
					"authorization", "cookie", "api_key", "bearer",
				},
			},
		},
		{
			name: "preserve custom max body size",
			config: Config{
				Debug: DebugConfig{
					MaxBodySize: 8192,
				},
			},
			want: Config{
				Debug: DebugConfig{
					MaxBodySize: 8192,
				},
				RedactKeywords: []string{
					"password", "token", "secret",
					"authorization", "cookie", "api_key", "bearer",
				},
			},
		},
		{
			name: "apply default redact keywords",
			config: Config{
				RedactKeywords: []string{},
			},
			want: Config{
				Debug: DebugConfig{
					MaxBodySize: 4096,
				},
				RedactKeywords: []string{
					"password", "token", "secret",
					"authorization", "cookie", "api_key", "bearer",
				},
			},
		},
		{
			name: "preserve custom redact keywords",
			config: Config{
				RedactKeywords: []string{"custom_secret"},
			},
			want: Config{
				Debug: DebugConfig{
					MaxBodySize: 4096,
				},
				RedactKeywords: []string{"custom_secret"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.config
			cfg.ApplyDefaults()

			if cfg.Debug.MaxBodySize != tt.want.Debug.MaxBodySize {
				t.Errorf("MaxBodySize = %d, want %d",
					cfg.Debug.MaxBodySize, tt.want.Debug.MaxBodySize)
			}

			if len(cfg.RedactKeywords) != len(tt.want.RedactKeywords) {
				t.Errorf("RedactKeywords length = %d, want %d",
					len(cfg.RedactKeywords), len(tt.want.RedactKeywords))
			}
		})
	}
}

func TestFeature_New(t *testing.T) {
	tests := []struct {
		name       string
		yamlConfig string
		wantErr    bool
	}{
		{
			name: "valid minimal config",
			yamlConfig: `
exclude_paths: ["/health"]
redact_keywords: ["password"]
debug:
  log_request_headers: true
  max_body_size: 4096
`,
			wantErr: false,
		},
		{
			name: "valid config with all fields",
			yamlConfig: `
exclude_paths: ["/health", "/metrics"]
redact_keywords: ["password", "token", "secret"]
debug:
  log_request_headers: true
  log_response_headers: false
  log_request_body: true
  log_response_body: false
  max_body_size: 8192
`,
			wantErr: false,
		},
		{
			name: "empty config uses defaults",
			yamlConfig: `
exclude_paths: []
redact_keywords: []
debug:
  max_body_size: 0
`,
			wantErr: false,
		},
		{
			name: "invalid max_body_size too small",
			yamlConfig: `
debug:
  max_body_size: 512
`,
			wantErr: true,
		},
		{
			name: "invalid max_body_size too large",
			yamlConfig: `
debug:
  max_body_size: 2097152
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var node yaml.Node
			if err := yaml.Unmarshal([]byte(tt.yamlConfig), &node); err != nil {
				t.Fatalf("Failed to parse YAML: %v", err)
			}

			featureCfg := config.FeatureConfig{
				Enabled: true,
				Config:  node,
			}

			_, err := New(featureCfg)

			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFeature_Middleware(t *testing.T) {
	yamlConfig := `
exclude_paths: ["/health"]
redact_keywords: ["password"]
debug:
  log_request_headers: true
  max_body_size: 4096
`

	var node yaml.Node
	if err := yaml.Unmarshal([]byte(yamlConfig), &node); err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	featureCfg := config.FeatureConfig{
		Enabled: true,
		Config:  node,
	}

	feature, err := New(featureCfg)
	if err != nil {
		t.Fatalf("Failed to create feature: %v", err)
	}

	middleware := feature.Middleware()
	if middleware == nil {
		t.Error("Expected middleware to be non-nil")
	}
}
