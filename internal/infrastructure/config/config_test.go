package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/infrastructure/config"
)

func TestDecode_SingleDatabase(t *testing.T) {
	// Arrange
	input := `
[[databases]]
name = "local"
db_path = "/tmp/example.sqlite"
`

	// Act
	cfg, err := config.Decode(strings.NewReader(input))

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cfg.Databases) != 1 {
		t.Fatalf("expected 1 database, got %d", len(cfg.Databases))
	}
	if cfg.Databases[0].Name != "local" {
		t.Fatalf("expected name %q, got %q", "local", cfg.Databases[0].Name)
	}
	if cfg.Databases[0].Path != "/tmp/example.sqlite" {
		t.Fatalf("expected path %q, got %q", "/tmp/example.sqlite", cfg.Databases[0].Path)
	}
}

func TestDecode_MultipleDatabases(t *testing.T) {
	// Arrange
	input := `
[[databases]]
name = "local"
db_path = "/tmp/example.sqlite"

[[databases]]
name = "analytics"
db_path = "/tmp/analytics.sqlite"
`

	// Act
	cfg, err := config.Decode(strings.NewReader(input))

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(cfg.Databases) != 2 {
		t.Fatalf("expected 2 databases, got %d", len(cfg.Databases))
	}
	if cfg.Databases[0].Name != "local" {
		t.Fatalf("expected name %q, got %q", "local", cfg.Databases[0].Name)
	}
	if cfg.Databases[0].Path != "/tmp/example.sqlite" {
		t.Fatalf("expected path %q, got %q", "/tmp/example.sqlite", cfg.Databases[0].Path)
	}
	if cfg.Databases[1].Name != "analytics" {
		t.Fatalf("expected name %q, got %q", "analytics", cfg.Databases[1].Name)
	}
	if cfg.Databases[1].Path != "/tmp/analytics.sqlite" {
		t.Fatalf("expected path %q, got %q", "/tmp/analytics.sqlite", cfg.Databases[1].Path)
	}
}

func TestDecode_MissingDatabaseSection(t *testing.T) {
	// Arrange
	input := `title = "dbc"`

	// Act
	_, err := config.Decode(strings.NewReader(input))

	// Assert
	if !errors.Is(err, config.ErrMissingDatabase) {
		t.Fatalf("expected error %v, got %v", config.ErrMissingDatabase, err)
	}
}

func TestDecode_EmptyDatabasesList(t *testing.T) {
	// Arrange
	input := `databases = []`

	// Act
	_, err := config.Decode(strings.NewReader(input))

	// Assert
	if !errors.Is(err, config.ErrMissingDatabase) {
		t.Fatalf("expected error %v, got %v", config.ErrMissingDatabase, err)
	}
}

func TestDecode_LegacyDatabaseSection(t *testing.T) {
	// Arrange
	input := `
[database]
name = "legacy"
db_path = "/tmp/example.sqlite"
`

	// Act
	_, err := config.Decode(strings.NewReader(input))

	// Assert
	if !errors.Is(err, config.ErrMissingDatabase) {
		t.Fatalf("expected error %v, got %v", config.ErrMissingDatabase, err)
	}
}

func TestDecode_MultipleDatabasesMissingName(t *testing.T) {
	// Arrange
	input := `
[[databases]]
db_path = "/tmp/example.sqlite"
`

	// Act
	_, err := config.Decode(strings.NewReader(input))

	// Assert
	if !errors.Is(err, config.ErrMissingDatabaseName) {
		t.Fatalf("expected error %v, got %v", config.ErrMissingDatabaseName, err)
	}
}

func TestDecode_MultipleDatabasesMissingPath(t *testing.T) {
	// Arrange
	input := `
[[databases]]
name = "local"
`

	// Act
	_, err := config.Decode(strings.NewReader(input))

	// Assert
	if !errors.Is(err, config.ErrMissingDatabasePath) {
		t.Fatalf("expected error %v, got %v", config.ErrMissingDatabasePath, err)
	}
}

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
