package httplog

// Config defines the HTTP logging feature configuration.
type Config struct {
	ExcludePaths   []string    `yaml:"exclude_paths" validate:"omitempty"`
	RedactKeywords []string    `yaml:"redact_keywords" validate:"omitempty"`
	Debug          DebugConfig `yaml:"debug"`
}

// DebugConfig controls debug-level logging (headers, bodies).
type DebugConfig struct {
	LogRequestHeaders  bool `yaml:"log_request_headers"`
	LogResponseHeaders bool `yaml:"log_response_headers"`
	LogRequestBody     bool `yaml:"log_request_body"`
	LogResponseBody    bool `yaml:"log_response_body"`
	MaxBodySize        int  `yaml:"max_body_size" validate:"omitempty,min=1024,max=1048576"` // 1KB-1MB
}

// ApplyDefaults sets default values for missing configuration.
func (c *Config) ApplyDefaults() {
	if c.Debug.MaxBodySize == 0 {
		c.Debug.MaxBodySize = 4096 // 4KB default
	}

	// Default redact keywords if none specified
	if len(c.RedactKeywords) == 0 {
		c.RedactKeywords = []string{
			"password",
			"token",
			"secret",
			"authorization",
			"cookie",
			"api_key",
			"bearer",
		}
	}
}
