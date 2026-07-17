package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	// нет файла и env → дефолт json
	dir := t.TempDir()
	t.Setenv("CONFIG_FILE", filepath.Join(dir, "absent.yaml"))
	t.Setenv("OUTPUT_FORMAT", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OutputFormat != "json" {
		t.Fatalf("OutputFormat: ожидали json, получили %q", cfg.OutputFormat)
	}
}

func TestLoad_FileAndEnvOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte("output_format: yaml\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONFIG_FILE", path)
	t.Setenv("OUTPUT_FORMAT", "text") // env бьёт файл
	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.OutputFormat != "text" {
		t.Fatalf("env override не сработал: %q", cfg.OutputFormat)
	}
}
