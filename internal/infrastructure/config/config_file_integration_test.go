package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mgierok/dbc/internal/infrastructure/config"
)

func TestLoadFile_LoadsConfigFromPath(t *testing.T) {
	// Arrange
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
[[databases]]
name = "local"
db_path = "/tmp/example.sqlite"
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}

	// Act
	cfg, err := config.LoadFile(path)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cfg.Databases) != 1 {
		t.Fatalf("expected one database, got %d", len(cfg.Databases))
	}
	if cfg.Databases[0].Name != "local" {
		t.Fatalf("expected name %q, got %q", "local", cfg.Databases[0].Name)
	}
	if cfg.Databases[0].Path != "/tmp/example.sqlite" {
		t.Fatalf("expected path %q, got %q", "/tmp/example.sqlite", cfg.Databases[0].Path)
	}
}

func TestLoadFile_ReturnsErrorWhenFileDoesNotExist(t *testing.T) {
	// Arrange
	missingPath := filepath.Join(t.TempDir(), "missing.toml")

	// Act
	_, err := config.LoadFile(missingPath)

	// Assert
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}
