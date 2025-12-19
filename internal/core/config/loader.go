package config

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/noyellowline/gokart/internal/core/errors"
	"gopkg.in/yaml.v3"
)

const DefaultConfigPath = "configs/proxy.yaml"

var validate = validator.New(validator.WithRequiredStructEnabled())

func Load(path string) (*Config, error) {
	if path == "" {
		path = DefaultConfigPath
		slog.Info("no config path provided, using default", "path", path)
	}

	slog.Info("loading config from file", "path", path)
	cfg, unexpanded, err := loadConfigFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to load config file: %w", err)
	}

	if len(unexpanded) > 0 {
		return nil, fmt.Errorf("required environment variables not set: %v", unexpanded)
	}

	slog.Info("config file loaded successfully", "path", path)

	if err := validate.Struct(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", errors.FormatValidation(err))
	}

	slog.Info("config validation passed")
	return cfg, nil
}

func loadConfigFile(path string) (*Config, []string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, err
	}

	expanded, unexpanded, err := expandTemplates(data)
	if err != nil {
		return nil, nil, fmt.Errorf("template expansion failed: %w", err)
	}

	if len(unexpanded) == 0 {
		slog.Info("all template variables expanded successfully")
	}

	var cfg Config
	if err := yaml.Unmarshal(expanded, &cfg); err != nil {
		return nil, nil, fmt.Errorf("YAML parsing failed: %w", err)
	}

	return &cfg, unexpanded, nil
}
