package httplog

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/noyellowline/gokart/internal/core/config"
)

// Feature: HTTP logging.
type Feature struct {
	config *Config
}

// New creates a new HTTP logging feature from configuration.
func New(featureCfg config.FeatureConfig) (*Feature, error) {
	var cfg Config

	// Decode YAML config
	if err := featureCfg.Config.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("failed to decode httplog config: %w", err)
	}

	// Apply defaults
	cfg.ApplyDefaults()

	// Validate config
	validate := validator.New()
	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("httplog config validation failed: %w", err)
	}

	return &Feature{config: &cfg}, nil
}

// Middleware returns the HTTP logging middleware.
func (f *Feature) Middleware() func(http.Handler) http.Handler {
	return Middleware(f.config)
}
