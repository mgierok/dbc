package config_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/mgierok/dbc/internal/application/port"
	"github.com/mgierok/dbc/internal/infrastructure/config"
)

func TestConfigStore_CRUDPersistence(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "config.toml")
	content := `
[[databases]]
name = "local"
db_path = "/tmp/local.sqlite"
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}
	store := config.NewStore(path)

	// Act
	err := store.Create(context.Background(), port.ConfigEntry{Name: "analytics", DBPath: "/tmp/analytics.sqlite"})
	if err != nil {
		t.Fatalf("expected no create error, got %v", err)
	}
	err = store.Update(context.Background(), 0, port.ConfigEntry{Name: "primary", DBPath: "/tmp/primary.sqlite"})
	if err != nil {
		t.Fatalf("expected no update error, got %v", err)
	}
	err = store.Delete(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no delete error, got %v", err)
	}
	entries, err := store.List(context.Background())

	// Assert
	if err != nil {
		t.Fatalf("expected no list error, got %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected one entry, got %d", len(entries))
	}
	if entries[0].Name != "primary" {
		t.Fatalf("expected name %q, got %q", "primary", entries[0].Name)
	}
	if entries[0].DBPath != "/tmp/primary.sqlite" {
		t.Fatalf("expected path %q, got %q", "/tmp/primary.sqlite", entries[0].DBPath)
	}
}

func TestConfigStore_CreateRejectsInvalidEntry(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "config.toml")
	content := `
[[databases]]
name = "local"
db_path = "/tmp/local.sqlite"
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}
	store := config.NewStore(path)

	// Act
	err := store.Create(context.Background(), port.ConfigEntry{Name: " ", DBPath: "/tmp/analytics.sqlite"})

	// Assert
	if !errors.Is(err, config.ErrMissingDatabaseName) {
		t.Fatalf("expected error %v, got %v", config.ErrMissingDatabaseName, err)
	}
}

func TestConfigStore_ListReturnsEmptyWhenFileDoesNotExist(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "missing.toml")
	store := config.NewStore(path)

	// Act
	entries, err := store.List(context.Background())

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty entries, got %d", len(entries))
	}
}

func TestConfigStore_ListReturnsEmptyWhenConfigHasNoDatabases(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("databases = []\n"), 0o600); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}
	store := config.NewStore(path)

	// Act
	entries, err := store.List(context.Background())

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty entries, got %d", len(entries))
	}
}

func TestConfigStore_ListReturnsErrorWhenConfigUsesUnknownShape(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "config.toml")
	content := `
[database]
name = "legacy"
db_path = "/tmp/legacy.sqlite"
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}
	store := config.NewStore(path)

	// Act
	_, err := store.List(context.Background())

	// Assert
	if err == nil {
		t.Fatal("expected malformed config error, got nil")
	}
}

func TestConfigStore_CreateCreatesConfigWhenFileDoesNotExist(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "missing.toml")
	store := config.NewStore(path)

	// Act
	err := store.Create(context.Background(), port.ConfigEntry{
		Name:   "local",
		DBPath: "/tmp/local.sqlite",
	})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	entries, listErr := store.List(context.Background())
	if listErr != nil {
		t.Fatalf("expected list without error, got %v", listErr)
	}
	if len(entries) != 1 {
		t.Fatalf("expected one entry, got %d", len(entries))
	}
	if entries[0].Name != "local" || entries[0].DBPath != "/tmp/local.sqlite" {
		t.Fatalf("unexpected entry: %#v", entries[0])
	}
}

func TestConfigStore_CreateCreatesFirstEntryWhenConfigHasNoDatabases(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte(""), 0o600); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}
	store := config.NewStore(path)

	// Act
	err := store.Create(context.Background(), port.ConfigEntry{
		Name:   "local",
		DBPath: "/tmp/local.sqlite",
	})

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	entries, listErr := store.List(context.Background())
	if listErr != nil {
		t.Fatalf("expected list without error, got %v", listErr)
	}
	if len(entries) != 1 {
		t.Fatalf("expected one entry, got %d", len(entries))
	}
	if entries[0].Name != "local" || entries[0].DBPath != "/tmp/local.sqlite" {
		t.Fatalf("unexpected entry: %#v", entries[0])
	}
}

func TestConfigStore_CreateReturnsErrorWhenConfigHasInvalidSyntax(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "config.toml")
	if err := os.WriteFile(path, []byte("[[databases]\n"), 0o600); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}
	store := config.NewStore(path)

	// Act
	err := store.Create(context.Background(), port.ConfigEntry{
		Name:   "local",
		DBPath: "/tmp/local.sqlite",
	})

	// Assert
	if err == nil {
		t.Fatal("expected error for invalid config syntax, got nil")
	}
}

func TestConfigStore_DeleteAllowsRemovingLastEntry(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "config.toml")
	content := `
[[databases]]
name = "local"
db_path = "/tmp/local.sqlite"
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}
	store := config.NewStore(path)

	// Act
	err := store.Delete(context.Background(), 0)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	entries, listErr := store.List(context.Background())
	if listErr != nil {
		t.Fatalf("expected list without error, got %v", listErr)
	}
	if len(entries) != 0 {
		t.Fatalf("expected zero entries, got %d", len(entries))
	}
}
