package config_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/infrastructure/config"
)

func TestDecode_ValidConfig(t *testing.T) {
	// Arrange
	input := `
[database]
name = "local"
db_path = "/tmp/example.sqlite"
`

	// Act
	cfg, err := config.Decode(strings.NewReader(input))

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Database == nil {
		t.Fatal("expected database config, got nil")
	}
	if cfg.Database.Name != "local" {
		t.Fatalf("expected name %q, got %q", "local", cfg.Database.Name)
	}
	if cfg.Database.Path != "/tmp/example.sqlite" {
		t.Fatalf("expected path %q, got %q", "/tmp/example.sqlite", cfg.Database.Path)
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

func TestDecode_MissingDatabaseName(t *testing.T) {
	// Arrange
	input := `
[database]
db_path = "/tmp/example.sqlite"
`

	// Act
	_, err := config.Decode(strings.NewReader(input))

	// Assert
	if !errors.Is(err, config.ErrMissingDatabaseName) {
		t.Fatalf("expected error %v, got %v", config.ErrMissingDatabaseName, err)
	}
}

func TestDecode_MissingDatabasePath(t *testing.T) {
	// Arrange
	input := `
[database]
name = "local"
`

	// Act
	_, err := config.Decode(strings.NewReader(input))

	// Assert
	if !errors.Is(err, config.ErrMissingDatabasePath) {
		t.Fatalf("expected error %v, got %v", config.ErrMissingDatabasePath, err)
	}
}
