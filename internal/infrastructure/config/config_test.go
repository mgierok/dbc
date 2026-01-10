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

func TestDatabaseList_SingleDatabase(t *testing.T) {
	// Arrange
	cfg := config.Config{
		Database: &config.DatabaseConfig{
			Name: "local",
			Path: "/tmp/example.sqlite",
		},
	}

	// Act
	databases, err := cfg.DatabaseList()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(databases) != 1 {
		t.Fatalf("expected 1 database, got %d", len(databases))
	}
	if databases[0].Name != "local" {
		t.Fatalf("expected name %q, got %q", "local", databases[0].Name)
	}
	if databases[0].Path != "/tmp/example.sqlite" {
		t.Fatalf("expected path %q, got %q", "/tmp/example.sqlite", databases[0].Path)
	}
}

func TestDatabaseList_MultipleDatabases(t *testing.T) {
	// Arrange
	cfg := config.Config{
		Databases: []config.DatabaseConfig{
			{Name: "local", Path: "/tmp/example.sqlite"},
			{Name: "analytics", Path: "/tmp/analytics.sqlite"},
		},
	}

	// Act
	databases, err := cfg.DatabaseList()

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(databases) != 2 {
		t.Fatalf("expected 2 databases, got %d", len(databases))
	}
	if databases[1].Name != "analytics" {
		t.Fatalf("expected name %q, got %q", "analytics", databases[1].Name)
	}
	if databases[1].Path != "/tmp/analytics.sqlite" {
		t.Fatalf("expected path %q, got %q", "/tmp/analytics.sqlite", databases[1].Path)
	}
}
