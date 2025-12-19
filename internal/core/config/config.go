package config

import (
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Core     CoreConfig               `yaml:"core" validate:"required"`
	Features map[string]FeatureConfig `yaml:"features"`
}

type CoreConfig struct {
	Server ServerConfig `yaml:"server" validate:"required"`
	Proxy  ProxyConfig  `yaml:"proxy" validate:"required"`
}

type ServerConfig struct {
	Addr           string        `yaml:"addr" validate:"required,tcp_addr"`
	ReadTimeout    time.Duration `yaml:"read_timeout" validate:"omitempty,gt=0"`
	WriteTimeout   time.Duration `yaml:"write_timeout" validate:"omitempty,gt=0"`
	IdleTimeout    time.Duration `yaml:"idle_timeout" validate:"omitempty,gt=0"`
	MaxHeaderBytes int           `yaml:"max_header_bytes" validate:"omitempty,min=1,max=104857600"`
}

type ProxyConfig struct {
	Target          string        `yaml:"target" validate:"required,url"`
	MaxIdleConns    int           `yaml:"max_idle_conns" validate:"omitempty,min=1,max=1000"`
	IdleConnTimeout time.Duration `yaml:"idle_conn_timeout" validate:"omitempty,gt=0"`
	FlushInterval   time.Duration `yaml:"flush_interval" validate:"omitempty,gt=0"`
}

type FeatureConfig struct {
	Enabled bool      `yaml:"enabled"`
	Config  yaml.Node `yaml:"config"`
}
