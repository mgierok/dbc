package config_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/application/port"
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

func TestDecode_EmptyDocument(t *testing.T) {
	// Arrange
	input := ``

	// Act
	_, err := config.Decode(strings.NewReader(input))

	// Assert
	if !errors.Is(err, config.ErrMissingDatabase) {
		t.Fatalf("expected error %v, got %v", config.ErrMissingDatabase, err)
	}
}

func TestDecode_UnknownTopLevelField(t *testing.T) {
	// Arrange
	input := `title = "dbc"`

	// Act
	_, err := config.Decode(strings.NewReader(input))

	// Assert
	if err == nil {
		t.Fatal("expected malformed config error, got nil")
	}
	if errors.Is(err, config.ErrMissingDatabase) {
		t.Fatalf("expected malformed config error, got empty-config error %v", err)
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
	if err == nil {
		t.Fatal("expected malformed config error, got nil")
	}
	if errors.Is(err, config.ErrMissingDatabase) {
		t.Fatalf("expected malformed config error, got empty-config error %v", err)
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

func TestResolvePathForOS_LinuxUsesHomeConfig(t *testing.T) {
	// Arrange
	home := "/home/tester"

	// Act
	path := config.ResolvePathForOS("linux", home, "")

	// Assert
	expected := filepath.Join(home, ".config", "dbc", "config.toml")
	if path != expected {
		t.Fatalf("expected %q, got %q", expected, path)
	}
}

func TestResolvePathForOS_MacOSUsesHomeConfig(t *testing.T) {
	// Arrange
	home := "/Users/tester"

	// Act
	path := config.ResolvePathForOS("darwin", home, "")

	// Assert
	expected := filepath.Join(home, ".config", "dbc", "config.toml")
	if path != expected {
		t.Fatalf("expected %q, got %q", expected, path)
	}
}

func TestResolvePathForOS_WindowsUsesAppData(t *testing.T) {
	// Arrange
	appData := "C:/Users/tester/AppData/Roaming"

	// Act
	path := config.ResolvePathForOS("windows", "C:/Users/tester", appData)

	// Assert
	expected := filepath.Join(appData, "dbc", "config.toml")
	if path != expected {
		t.Fatalf("expected %q, got %q", expected, path)
	}
}

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
	if errors.Is(err, config.ErrMissingDatabase) {
		t.Fatalf("expected malformed config error, got empty-config error %v", err)
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
