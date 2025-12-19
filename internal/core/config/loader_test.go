package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ValidConfig(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "test.yaml")
	yaml := `
core:
  server:
    addr: "0.0.0.0:8080"
    read_timeout: 10s
    write_timeout: 10s
    idle_timeout: 120s
  proxy:
    target: "http://localhost:3000"
    max_idle_conns: 100
    idle_conn_timeout: 90s
`
	if err := os.WriteFile(tmpFile, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Core.Server.Addr != "0.0.0.0:8080" {
		t.Errorf("got addr %s, want 0.0.0.0:8080", cfg.Core.Server.Addr)
	}
}

func TestLoad_WithEnvVars(t *testing.T) {
	os.Setenv("TEST_PORT", "9000")
	defer os.Unsetenv("TEST_PORT")

	tmpFile := filepath.Join(t.TempDir(), "test.yaml")
	yaml := `
core:
  server:
    addr: "0.0.0.0:${TEST_PORT}"
    read_timeout: 10s
    write_timeout: 10s
    idle_timeout: 120s
  proxy:
    target: "http://localhost:3000"
`
	if err := os.WriteFile(tmpFile, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(tmpFile)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Core.Server.Addr != "0.0.0.0:9000" {
		t.Errorf("got addr %s, want 0.0.0.0:9000", cfg.Core.Server.Addr)
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/file.yaml")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "invalid.yaml")
	if err := os.WriteFile(tmpFile, []byte("invalid: yaml: content:"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(tmpFile)
	if err == nil {
		t.Fatal("expected error for invalid YAML, got nil")
	}
}

func TestLoad_ValidationFails(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "invalid.yaml")
	yaml := `
core:
  server:
    addr: "invalid_addr"
    read_timeout: 10s
    write_timeout: 10s
    idle_timeout: 120s
  proxy:
    target: "not_a_url"
`
	if err := os.WriteFile(tmpFile, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(tmpFile)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestLoad_MissingRequiredEnvVar(t *testing.T) {
	os.Unsetenv("REQUIRED_VAR")

	tmpFile := filepath.Join(t.TempDir(), "test.yaml")
	yaml := `
core:
  server:
    addr: "${REQUIRED_VAR}"
    read_timeout: 10s
    write_timeout: 10s
    idle_timeout: 120s
  proxy:
    target: "http://localhost:3000"
`
	if err := os.WriteFile(tmpFile, []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(tmpFile)
	if err == nil {
		t.Fatal("expected error for missing env var, got nil")
	}
}
