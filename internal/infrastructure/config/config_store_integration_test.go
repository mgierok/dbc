package config_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mgierok/dbc/internal/application/port"
	"github.com/mgierok/dbc/internal/infrastructure/config"
)

func TestConfigStore_CRUDPersistence(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "config.json")
	content := `{"databases":[{"name":"local","db_path":"/tmp/local.sqlite"}]}`
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
	path := filepath.Join(t.TempDir(), "config.json")
	content := `{"databases":[{"name":"local","db_path":"/tmp/local.sqlite"}]}`
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
	path := filepath.Join(t.TempDir(), "missing.json")
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
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"databases":[]}`), 0o600); err != nil {
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
	path := filepath.Join(t.TempDir(), "config.json")
	content := `{"database":{"name":"legacy","db_path":"/tmp/legacy.sqlite"}}`
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
	path := filepath.Join(t.TempDir(), "missing.json")
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
	path := filepath.Join(t.TempDir(), "config.json")
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
	path := filepath.Join(t.TempDir(), "config.json")
	if err := os.WriteFile(path, []byte(`{"databases":[`), 0o600); err != nil {
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
	path := filepath.Join(t.TempDir(), "config.json")
	content := `{"databases":[{"name":"local","db_path":"/tmp/local.sqlite"}]}`
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

func TestConfigStore_CreateReturnsErrorWhenSerializedConfigExceedsSizeLimit(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "config.json")
	store := config.NewStore(path)
	entry := configEntryForSerializedSize(t, config.Config{}, 1<<20, false)

	// Act
	err := store.Create(context.Background(), entry)

	// Assert
	if !errors.Is(err, config.ErrConfigTooLarge) {
		t.Fatalf("expected error %v, got %v", config.ErrConfigTooLarge, err)
	}
	if _, statErr := os.Stat(path); !errors.Is(statErr, os.ErrNotExist) {
		t.Fatalf("expected config file to remain absent, got stat error %v", statErr)
	}
}

func TestConfigStore_UpdateReturnsErrorWhenSerializedConfigExceedsSizeLimit(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "config.json")
	initialContent := `{"databases":[{"name":"local","db_path":"/tmp/local.sqlite"}]}`
	if err := os.WriteFile(path, []byte(initialContent), 0o600); err != nil {
		t.Fatalf("failed to write temp config file: %v", err)
	}
	store := config.NewStore(path)
	baseConfig := config.Config{
		Databases: []config.DatabaseConfig{
			{Name: "local", Path: "/tmp/local.sqlite"},
		},
	}
	entry := replacementConfigEntryForSerializedSize(t, baseConfig, 0, 1<<20, false)

	// Act
	err := store.Update(context.Background(), 0, entry)

	// Assert
	if !errors.Is(err, config.ErrConfigTooLarge) {
		t.Fatalf("expected error %v, got %v", config.ErrConfigTooLarge, err)
	}
	cfg, loadErr := config.LoadFile(path)
	if loadErr != nil {
		t.Fatalf("expected previous config to remain readable, got %v", loadErr)
	}
	if len(cfg.Databases) != 1 {
		t.Fatalf("expected one database, got %d", len(cfg.Databases))
	}
	if cfg.Databases[0].Name != "local" || cfg.Databases[0].Path != "/tmp/local.sqlite" {
		t.Fatalf("expected previous config content to remain unchanged, got %#v", cfg.Databases[0])
	}
}

func TestConfigStore_CreateAllowsSerializedConfigAtSizeLimit(t *testing.T) {
	// Arrange
	path := filepath.Join(t.TempDir(), "config.json")
	store := config.NewStore(path)
	entry := configEntryForSerializedSize(t, config.Config{}, 1<<20, true)

	// Act
	err := store.Create(context.Background(), entry)

	// Assert
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	info, statErr := os.Stat(path)
	if statErr != nil {
		t.Fatalf("failed to stat written config file: %v", statErr)
	}
	if info.Size() != 1<<20 {
		t.Fatalf("expected written file size %d, got %d", 1<<20, info.Size())
	}
	cfg, loadErr := config.LoadFile(path)
	if loadErr != nil {
		t.Fatalf("expected written config to remain readable, got %v", loadErr)
	}
	if len(cfg.Databases) != 1 {
		t.Fatalf("expected one database, got %d", len(cfg.Databases))
	}
}

func configEntryForSerializedSize(t *testing.T, base config.Config, targetSize int, exact bool) port.ConfigEntry {
	t.Helper()

	const dbPath = "/tmp/local.sqlite"

	entry := port.ConfigEntry{
		Name:   "a",
		DBPath: dbPath,
	}

	currentSize := serializedConfigSize(t, appendConfigEntry(base, entry))
	if currentSize > targetSize {
		t.Fatalf("base config already exceeds target size: %d > %d", currentSize, targetSize)
	}

	padding := targetSize - currentSize
	if !exact {
		padding++
	}
	entry.Name += strings.Repeat("a", padding)

	finalSize := serializedConfigSize(t, appendConfigEntry(base, entry))
	if exact && finalSize != targetSize {
		t.Fatalf("expected serialized size %d, got %d", targetSize, finalSize)
	}
	if !exact && finalSize <= targetSize {
		t.Fatalf("expected serialized size above %d, got %d", targetSize, finalSize)
	}

	return entry
}

func appendConfigEntry(base config.Config, entry port.ConfigEntry) config.Config {
	cfg := config.Config{
		Databases: append([]config.DatabaseConfig(nil), base.Databases...),
	}
	cfg.Databases = append(cfg.Databases, config.DatabaseConfig{
		Name: entry.Name,
		Path: entry.DBPath,
	})
	return cfg
}

func replaceConfigEntry(base config.Config, index int, entry port.ConfigEntry) config.Config {
	cfg := config.Config{
		Databases: append([]config.DatabaseConfig(nil), base.Databases...),
	}
	cfg.Databases[index] = config.DatabaseConfig{
		Name: entry.Name,
		Path: entry.DBPath,
	}
	return cfg
}

func replacementConfigEntryForSerializedSize(t *testing.T, base config.Config, index int, targetSize int, exact bool) port.ConfigEntry {
	t.Helper()

	const dbPath = "/tmp/local.sqlite"

	entry := port.ConfigEntry{
		Name:   "a",
		DBPath: dbPath,
	}

	currentSize := serializedConfigSize(t, replaceConfigEntry(base, index, entry))
	if currentSize > targetSize {
		t.Fatalf("base config already exceeds target size: %d > %d", currentSize, targetSize)
	}

	padding := targetSize - currentSize
	if !exact {
		padding++
	}
	entry.Name += strings.Repeat("a", padding)

	finalSize := serializedConfigSize(t, replaceConfigEntry(base, index, entry))
	if exact && finalSize != targetSize {
		t.Fatalf("expected serialized size %d, got %d", targetSize, finalSize)
	}
	if !exact && finalSize <= targetSize {
		t.Fatalf("expected serialized size above %d, got %d", targetSize, finalSize)
	}

	return entry
}

func serializedConfigSize(t *testing.T, cfg config.Config) int {
	t.Helper()

	content, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("failed to serialize config: %v", err)
	}
	return len(content) + len("\n")
}
