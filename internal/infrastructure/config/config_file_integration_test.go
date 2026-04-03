package config_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/infrastructure/config"
)

func TestLoadFile_LoadsConfigFromPath(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "config.json")
	writeConfigFile(t, path, `{"databases":[{"name":"local","db_path":"/tmp/example.sqlite"}]}`)

	// Act
	got, err := config.LoadFile(path)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	assertDatabaseConfigs(t, got.Databases, []config.DatabaseConfig{
		{Name: "local", Path: "/tmp/example.sqlite"},
	})
}

func TestLoadFile_ReturnsEmptyConfigForTrimmedEmptyFile(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "config.json")
	writeConfigFile(t, path, " \n\t ")

	// Act
	got, err := config.LoadFile(path)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	assertDatabaseConfigs(t, got.Databases, nil)
}

func TestLoadFile_ReturnsErrorWhenFileDoesNotExist(t *testing.T) {
	// Arrange
	missingPath := filepath.Join(t.TempDir(), "missing.json")

	// Act
	_, err := config.LoadFile(missingPath)

	// Assert
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoadFile_ReturnsErrorWhenConfigExceedsSizeLimit(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "config.json")
	writeConfigFile(t, path, strings.Repeat(" ", (1<<20)+1))

	// Act
	_, err := config.LoadFile(path)

	// Assert
	if !errors.Is(err, config.ErrConfigTooLarge) {
		t.Fatalf("expected error %v, got %v", config.ErrConfigTooLarge, err)
	}
}

func writeConfigFile(t *testing.T, path string, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("failed to create config directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}
}
